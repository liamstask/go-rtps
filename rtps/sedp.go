package rtps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// sedp: Simple Endpoint Discovery Protocol

type sedpTopicInfo struct {
	guid      GUID
	topicName string
	typeName  string
}

type SEDP struct {
	lastBcast            time.Time
	subPub               *Pub // SEDP subscription publication
	pubPub               *Pub // SEDP publication publication
	subWriterDataSubmsgs []*submsgData
	pubWriterDataSubmsgs []*submsgData
	subWriterDataBuf     *bytes.Buffer //[SEDP_MSG_BUF_LEN * FRUDP_MAX_SUBS];
	pubWriterDataBuf     *bytes.Buffer //[SEDP_MSG_BUF_LEN * FRUDP_MAX_PUBS];
	topicInfo            sedpTopicInfo
}

func (s *SEDP) init() {
	// no topic, type name
	s.subPub = newPub("", "", SEDPSubWriterID, s.subWriterDataSubmsgs)

	// no topic, type name
	s.pubPub = newPub("", "", SEDPPubWriterID, s.pubWriterDataSubmsgs)

	// subscribe to the subscriber announcers
	defaultSession.addSub(&Sub{
		topicName: "",
		typeName:  "",
		readerEID: SEDPSubReaderID,
		dataCB:    s.rxSubData,
		reliable:  true,
	})

	// subscribe to the publisher announcers
	defaultSession.addSub(&Sub{
		topicName: "",
		typeName:  "",
		readerEID: SEDPPubReaderID,
		dataCB:    s.rxPubData,
		reliable:  true,
	})
}

func (s *SEDP) start() {
	// go through and send SEDP messages for all of our subscriptions
	// user subs have topic names

	for _, sub := range defaultSession.subs {
		if sub.topicName != "" {
			println("sending SEDP message about subscription:", sub.topicName)
			s.publishSub(sub)
		}
	}

	// now go through and send SEDP messages for all our publications
	for _, pub := range defaultSession.pubs {
		if pub.topicName != "" {
			println("sending SEDP message about publication:", pub.topicName)
			s.publishPub(pub)
		}
	}

	go func() {
		c := time.Tick(1 * time.Second)
		for range c {
			s.bcast()
		}
	}()
}

func (s *SEDP) bcast() {
	// todo
}

func (s *SEDP) rxPubData(r *receiver, submsg *subMsg, scheme uint16, b []byte) {
	s.rxPubsubData(r, submsg, scheme, b, true)
}

func (s *SEDP) rxSubData(r *receiver, submsg *subMsg, scheme uint16, b []byte) {
	s.rxPubsubData(r, submsg, scheme, b, false)
}

func (s *SEDP) addBuiltinEndpoints(part *Participant) {

	// this reads the remote peer's publications
	defaultSession.addReader(&Reader{
		writerGUID:  GUID{eid: SEDPPubWriterID, prefix: part.guidPrefix},
		readerEID:   SEDPPubReaderID,
		maxRxSeqNum: 0,
		dataCB:      s.rxPubData,
		reliable:    true,
	})

	// this reads the remote peer's subscriptions
	defaultSession.addReader(&Reader{
		writerGUID:  GUID{eid: SEDPSubWriterID, prefix: part.guidPrefix},
		readerEID:   SEDPSubReaderID,
		maxRxSeqNum: 0,
		dataCB:      s.rxSubData,
		reliable:    true,
	})

	// blast our SEDP data at this participant
	s.sendSedpMsgs(part)
}

func (s *SEDP) sendSedpMsgs(part *Participant) {
	defaultSession.spdp.bcast() // xxx: is this right?

	// we have just found out about a new participant. blast our SEDP messages
	// at it to help it join quickly

	// first, send the publications

	if s.pubPub.nextSubmsgIdx != 0 {
		var msgbuf bytes.Buffer

		tsSubmsg := newTsSubMsg(time.Now(), binary.LittleEndian)
		tsSubmsg.WriteTo(&msgbuf)

		for i := uint32(0); i < s.pubPub.nextSubmsgIdx; i++ {
			// todo: make sure we don't overflow a single ethernet frame
			pubSubmsg := s.pubPub.dataSubmsgs[i]
			pubSubmsg.WriteTo(&msgbuf)
			fmt.Printf("catchup SEDP msg %d addressed to reader EID 0x%08x\r\n", i, pubSubmsg.readerID)
		}
		hb := submsgHeartbeat{
			hdr: submsgHeader{
				id:    SUBMSG_ID_HEARTBEAT,
				flags: 0x3, // todo: spell this out
				sz:    28,
			},
			readerEID:   s.pubPub.dataSubmsgs[0].readerID,
			writerEID:   s.pubPub.dataSubmsgs[0].writerID,
			firstSeqNum: 1, // todo
			lastSeqNum:  newSeqNum(0, s.pubPub.nextSubmsgIdx),
			count:       0,
		}
		hb.WriteTo(&msgbuf)

		addr := part.metaUcastLoc.addrStr()
		fmt.Printf("sending %d bytes of SEDP catchup messages to %s\n", msgbuf.Len(), addr)
		if err := udpTXStr(msgbuf.Bytes(), addr); err != nil {
			println("couldn't transmit SPDP broadcast message:", err)
		}
	} else {
		println("no SEDP pub data to send to new participant")
	}
}

func (s *SEDP) rxPubsubData(r *receiver, submsg *subMsg, scheme uint16, b []byte, isPub bool) {
	if scheme != SCHEME_PL_CDR_LE {
		// FREERTPS_ERROR("expected sedp data to be PL_CDR_LE. bailing...\r\n");
		return
	}

	// memset(&s.topicInfo, 0, sizeof(sedp_topic_info_t));

	plist, _, err := newParamList(submsg.bin, b)
	if err != nil {
		// XXX: report
		return
	}

	for _, p := range plist {

		switch p.pid {
		case PID_ENDPOINT_GUID:
			// copy(s.topicInfo.guid[:], p.value)
			//if (guid.entity_id.u == 0x03010000)
			//  printf("found entity 0x103\n");

		case PID_TOPIC_NAME:
			if str, err := p.valToString(submsg.bin); err == nil {
				s.topicInfo.topicName = str
				println("    topic name:", str)
			} else {
				println("    couldn't parse topic name")
			}

		case PID_TYPE_NAME:
			if str, err := p.valToString(submsg.bin); err == nil {
				s.topicInfo.typeName = str
				println("    type name:", str)
			} else {
				println("    couldn't parse topic type")
			}

		case PID_RELIABILITY:
			if qos, err := newQosReliabilityFromBytes(submsg.bin, p.value); err == nil {
				if qos.kind == QOS_RELIABILITY_KIND_BEST_EFFORT {
					println("    reliability QoS: [best-effort]")
				} else if qos.kind == QOS_RELIABILITY_KIND_RELIABLE {
					println("    reliability QoS: [reliable]")
				} else {
					println("unhandled reliability kind:", qos.kind)
				}
			}

		case PID_HISTORY:
			if qos, err := newQosHistoryFromBytes(submsg.bin, p.value); err == nil {
				if qos.kind == QOS_HISTORY_KIND_KEEP_LAST {
					println("    history QoS: keep last", qos.depth)
				} else if qos.kind == QOS_HISTORY_KIND_KEEP_ALL {
					println("    history QoS: [keep all]")
				} else {
					println("unhandled history kind:", qos.kind)
				}
			}

		case PID_TRANSPORT_PRIORITY:
			if len(p.value) >= 4 {
				prio := submsg.bin.Uint32(b[0:])
				println("    transport priority:", prio)
			}
		}
	}

	// // make sure we have received all necessary parameters
	// if s.topicInfo.typeName == "" ||
	// 	s.topicInfo.topicName == "" ||
	// 	frudp_guid_identical(&s.topicInfo.guid, &g_frudp_guid_unknown) {
	// 	println("insufficient SEDP information")
	// 	return
	// }
	//
	// if isPub { // this is information about someone else's publication
	// 	frudp_sedp_rx_pub_info(s.topicInfo)
	// } else { // this is information about someone else's subscription
	// 	frudp_sedp_rx_sub_info(s.topicInfo)
	// }
}

func (s *SEDP) publishSub(sub *Sub) {
	if s.subPub == nil {
		println("woah there partner. you need to call frudp_part_create()")
		return
	}
	println("sedp_publish_sub:", sub.topicName)
	s.publish(sub.topicName, sub.typeName, s.subPub, sub.readerEID)
}

func (s *SEDP) publishPub(pub *Pub) {
	if s.pubPub == nil {
		println("woah there partner. you need to call frudp_part_create()")
		return
	}
	println("sedp_publish_pub:", pub.topicName)
	s.publish(pub.topicName, pub.typeName, s.pubPub, pub.writerEID)
}

func (s *SEDP) publish(topicName, typeName string, pub *Pub, eid EntityID) {
	// first make sure we have an spdp packet out first
	// printf("sedp publish [%s] via SEDP EID 0x%08x\r\n", topic_name, (unsigned)freertps_htonl(pub.writer_eid.u));
	// frudp_submsg_data_t *d = (frudp_submsg_data_t *)g_sedp_msg_buf;

	// data submessage
	dataSubmsg := submsgData{
		hdr: submsgHeader{
			id:    SUBMSG_ID_DATA,
			flags: FRUDP_FLAGS_LITTLE_ENDIAN | FRUDP_FLAGS_DATA_PRESENT,
		},
		extraflags:        0,
		octetsToInlineQos: 16,
		readerID:          SEDPSubReaderID,
		writerID:          SEDPSubWriterID,
		writerSeqNum:      0,
	}
	dataSubmsg.hdr.sz = 20

	var submsgBuf bytes.Buffer

	/////////////////////////////////////////////////////////////
	scheme := encapsulationScheme{
		scheme:  uint16(SCHEME_PL_CDR_LE),
		options: 0,
	}
	scheme.WriteTo(&submsgBuf)

	/////////////////////////////////////////////////////////////

	paramList := paramListItem{
		pid:   PID_PROTOCOL_VERSION,
		value: []byte{2, 1, 0, 0},
	}
	paramList.WriteTo(&submsgBuf)

	/////////////////////////////////////////////////////////////
	paramListNext := paramListItem{
		pid:   PID_VENDOR_ID,
		value: []byte{(MY_RTPS_VENDOR_ID >> 8) & 0xff, MY_RTPS_VENDOR_ID & 0xff, 0, 0},
	}
	paramListNext.WriteTo(&submsgBuf)

	/////////////////////////////////////////////////////////////
	epGUID := paramListItem{
		pid:   PID_ENDPOINT_GUID,
		value: defaultUDPConfig.guidPrefix[:],
	}
	epGUID.WriteTo(&submsgBuf)
	// frudp_guid_t guid;
	// guid.prefix = g_frudp_config.guid_prefix;
	// guid.eid = eid;
	// memcpy(param.value, &guid, 16);
	//printf("reader_guid = 0x%08x\n", htonl(reader_guid.entity_id.u));

	/////////////////////////////////////////////////////////////
	if topicName != "" {
		nameParam := paramListItem{
			pid:   PID_TOPIC_NAME,
			value: packParamString(binary.LittleEndian, topicName),
		}
		nameParam.WriteTo(&submsgBuf)
	}

	if typeName != "" {
		typeParam := paramListItem{
			pid:   PID_TYPE_NAME,
			value: packParamString(binary.LittleEndian, typeName),
		}
		typeParam.WriteTo(&submsgBuf)
	}
	/////////////////////////////////////////////////////////////
	// todo: follow the "reliable" flag in the subscription structure
	// reliability := qosReliability{
	//     kind:            QOS_RELIABILITY_KIND_BEST_EFFORT,
	//     maxBlockingTime: duration{sec: 0, nanosec: 0x19999999},
	// }
	qosParam := paramListItem{
		pid: PID_RELIABILITY,
		// xxx: value:,
	}
	qosParam.WriteTo(&submsgBuf)

	/////////////////////////////////////////////////////////////
	// XXX: PRESENTATION

	/////////////////////////////////////////////////////////////
	sentinel := paramListItem{pid: PID_RELIABILITY}
	sentinel.WriteTo(&submsgBuf)

	// d.header.len = param.value - 4 - (uint8_t *)&d.extraflags;
	udpPublish(pub, &dataSubmsg) // this will be either on the sub or pub publisher
}
