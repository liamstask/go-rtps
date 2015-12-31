package rtps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

// receiver is used to dispatch all the submsgs within a msg
// lifetime is a single msg
type receiver struct {
	srcProtoVer   ProtoVersion
	srcVID        VendorID
	srcGUIDPrefix GUIDPrefix
	dstGUIDPrefix GUIDPrefix
	// unicastReplyLocatorList   locator
	// multicastReplyLocatorList locator
	haveTimestamp bool
	timestamp     time.Time
}

// called when a new packet has been received
// parse packet, and handle submessages
func rxdispatch(b []byte) {
	hdr, err := newHeaderFromBytes(b)
	if err != nil {
		println("err header:", err)
		return
	}

	if Magic != hdr.magic {
		println("no magic here")
		return
	}

	if hdr.protoVer.major < MY_RTPS_VERSION_MAJOR {
		println("version too old")
		return
	}

	// don't process our own messages
	// xxx: why do we get these? better routing needed?
	if bytes.Equal(hdr.guidPrefix, defaultUDPConfig.guidPrefix) {
		return
	}

	// fmt.Printf("version: %d.%d vendor: 0x%x, guid: %v\n", hdr.protoVer.major, hdr.protoVer.minor, hdr.vid, hdr.guidPrefix.String())

	rxer := receiver{
		srcProtoVer:   hdr.protoVer,
		srcVID:        hdr.vid,
		srcGUIDPrefix: hdr.guidPrefix,
	}

	submsgbuf := b[8+len(hdr.guidPrefix):]

	for len(submsgbuf) >= 4 {
		submsg, err := newSubMsgFromBytes(submsgbuf)
		if err != nil {
			println("newSubMsgFromBytes:", err.Error())
			break
		}
		rxer.handleSubMsg(submsg)
		submsgbuf = submsgbuf[4+submsg.hdr.sz:]
	}

	if len(submsgbuf) > 0 {
		println("oops, done processing with", len(submsgbuf), "bytes remaining")
	}
}

func (r *receiver) handleSubMsg(sm *subMsg) {
	switch sm.hdr.id {
	case SUBMSG_ID_PAD:
		fmt.Println("SUBMSG_ID_PAD")

	case SUBMSG_ID_ACKNACK:
		r.rxAckNack(sm)

	case SUBMSG_ID_HEARTBEAT:
		r.rxHeartbeat(sm)

	case SUBMSG_ID_GAP:
		fmt.Println("SUBMSG_ID_GAP")

	case SUBMSG_ID_INFO_TS:
		r.rxInfoTS(sm)

	case SUBMSG_ID_INFO_SRC:
		r.rxInfoSrc(sm)

	case SUBMSG_ID_INFO_REPLY_IP4:
		fmt.Println("SUBMSG_ID_INFO_REPLY_IP4")

	case SUBMSG_ID_INFO_DST:
		r.rxInfoDst(sm)

	case SUBMSG_ID_INFO_REPLY:
		fmt.Println("SUBMSG_ID_INFO_REPLY")

	case SUBMSG_ID_NACK_FRAG:
		fmt.Println("SUBMSG_ID_NACK_FRAG")

	case SUBMSG_ID_HEARTBEAT_FRAG:
		fmt.Println("SUBMSG_ID_HEARTBEAT_FRAG")

	case SUBMSG_ID_DATA:
		r.rxData(sm)

	case SUBMSG_ID_DATA_FRAG:
		fmt.Println("SUBMSG_ID_DATA_FRAG")

	default:
		fmt.Println("**** SUBMSG_ID_UNKNOWN")
	}
}

// handler for SUBMSG_ID_INFO_TS submessages
func (r *receiver) rxInfoTS(sm *subMsg) {
	invalidate := sm.hdr.flags&FLAGS_INFOTS_INVALIDATE != 0
	if invalidate {
		r.haveTimestamp = false
		r.timestamp = timeInvalid
		println("INFO_TS: invalidate")
	} else {
		var err error
		if r.timestamp, err = timeFromBytes(sm.bin, sm.data); err == nil {
			r.haveTimestamp = true
			// fmt.Printf("INFO_TS: %v now: %v\n", r.timestamp, time.Now().UTC())
		}
	}
}

// handler for SUBMSG_ID_INFO_SRC submessages
func (r *receiver) rxInfoSrc(sm *subMsg) {

	is := submsgInfoSrc{
		// unused     uint32
		version:    ProtoVersion{sm.data[4], sm.data[5]},
		vid:        VendorID(binary.BigEndian.Uint16(sm.data[6:])),
		guidPrefix: sm.data[8 : 8+UDPGuidPrefixLen],
	}

	r.srcGUIDPrefix = is.guidPrefix
	r.srcProtoVer = is.version
	r.srcVID = is.vid
	fmt.Printf("INFOSRC: %s vendor 0x%x\n", r.srcGUIDPrefix.String(), r.srcVID)
}

// handler for SUBMSG_ID_INFO_DST submessages
func (r *receiver) rxInfoDst(sm *subMsg) {
	// only element in submsgInfoDest is the prefix
	if len(sm.data) == UDPGuidPrefixLen {
		r.dstGUIDPrefix = sm.data
	}
	fmt.Printf("INFO_DST (14): %s\n", r.dstGUIDPrefix.String())
}

// handler for SUBMSG_ID_DATA submessages
func (r *receiver) rxData(sm *subMsg) {
	inlineQoS := sm.hdr.flags&FLAGS_DATA_INLINE_QOS != 0
	//d := sm.hdr.flags & FLAGS_DATA_DATAFLAG
	keyed := sm.hdr.flags&FLAGS_DATA_KEYFLAG != 0
	if keyed {
		println("how to keyed data")
		return
	}

	// additional data-specific header info
	smd := submsgData{
		extraflags:        sm.bin.Uint16(sm.data[0:]),
		octetsToInlineQos: sm.bin.Uint16(sm.data[2:]),
		readerID:          EntityID(sm.bin.Uint32(sm.data[4:])),
		writerID:          EntityID(sm.bin.Uint32(sm.data[8:])),
		writerSeqNum:      newSeqNum(sm.bin.Uint32(sm.data[12:]), sm.bin.Uint32(sm.data[16:])),
		data:              sm.data[20:],
	}

	// fmt.Printf("  data: xflags 0x%x, octets to qos %d, 0x%x 0x%x, inlineQoS? %v\n", smd.extraflags, smd.octetsToInlineQos, smd.readerID, smd.writerID, inlineQoS)

	b := smd.data

	// parse and apply QoS parameters
	if inlineQoS {
		b = sm.data[4+smd.octetsToInlineQos:]
		_, n, err := newParamList(sm.bin, b)
		if err != nil {
			// XXX: report
			return
		}
		// for _, p := range plist {
		// 	// XXX: apply p
		// }
		b = b[n:]
	}

	es := newSchemeFromBytes(sm.bin, b)
	switch es.scheme {
	case SCHEME_CDR_LE, SCHEME_PL_CDR_LE:
	default:
		println("scheme unknown:", es.scheme)
	}
	b = b[4:]

	writerGUID := GUID{
		prefix: r.srcGUIDPrefix,
		eid:    smd.writerID,
	}

	// spin through subscriptions and see if anyone is listening
	matches := 0
	for _, rdr := range defaultSession.readers {
		if smd.Matches(rdr) || writerGUID.Equal(&rdr.writerGUID) {
			matches += 1
			// update the max-received sequence number counter
			if smd.writerSeqNum > rdr.maxRxSeqNum {
				rdr.maxRxSeqNum = smd.writerSeqNum
			}
			if rdr.dataCB != nil {
				rdr.dataCB(r, sm, es.scheme, b)
			}
			if rdr.msgCB != nil {
				rdr.msgCB(b)
			}
		}
	}

	if matches == 0 {
		fmt.Printf("    couldn't find a matched reader for this DATA: %s (0x%x)\n", writerGUID.prefix.String(), writerGUID.eid)
		println("    available readers:")
		for _, rdr := range defaultSession.readers {
			fmt.Printf("      writer = %s => 0x%x\n", rdr.writerGUID.prefix.String(), rdr.readerEID)
		}
	}
}

// handler for SUBMSG_ID_ACKNACK submessages
func (r *receiver) rxAckNack(sm *subMsg) {
	an := submsgAckNack{
		readerEID: EntityID(sm.bin.Uint32(sm.data[0:])),
		writerEID: EntityID(sm.bin.Uint32(sm.data[4:])),
		readerSNState: SeqNumSet{
			bitmapBase: newSeqNum(sm.bin.Uint32(sm.data[8:]), sm.bin.Uint32(sm.data[12:])),
			numBits:    sm.bin.Uint32(sm.data[16:]),
		},
	}

	fmt.Printf("ACKNACK: %d . %d\n", an.readerSNState.bitmapBase, an.readerSNState.numBits)

	if pub := defaultSession.pubWithWriterID(an.writerEID); pub != nil {
		pub.rxAckNack(&an, r.srcGUIDPrefix)
	} else {
		fmt.Printf("couldn't find pub for writer id 0x%08x\n", an.writerEID)
	}
}

// handler for SUBMSG_ID_HEARTBEAT submessages
func (r *receiver) rxHeartbeat(sm *subMsg) {

	final := sm.hdr.flags&FLAGS_HEARTBEAT_FLAG_FINAL != 0
	// lively := sm.hdr.flags&FLAGS_HEARTBEAT_FLAG_LIVELINESS != 0

	hb, err := newHeartbeatFromBytes(sm.data)
	if err != nil {
		println("newHeartbeatFromBytes:", err.Error())
		return
	}
	writerGUID := GUID{eid: hb.writerEID, prefix: r.srcGUIDPrefix}

	fmt.Printf("  HEARTBEAT %s", writerGUID.prefix.String())
	fmt.Printf(" => 0x%08x  %d.%d\n", hb.readerEID, hb.firstSeqNum, hb.lastSeqNum)

	var match *Reader
	// spin through subscriptions and see if we've already matched a reader

	for _, rdr := range defaultSession.readers {
		if writerGUID.Equal(&rdr.writerGUID) &&
			(hb.readerEID == rdr.readerEID || hb.readerEID == 0) {
			match = rdr
			break
		}
	}

	// else, if we have a subscription for this, initialize a reader
	if match == nil {
		for _, sub := range defaultSession.subs {
			if sub.readerEID == hb.readerEID {
				r := &Reader{
					reliable:    sub.reliable,
					readerEID:   hb.readerEID,
					maxRxSeqNum: 0,
					dataCB:      sub.dataCB,
					msgCB:       sub.msgCB,
					writerGUID:  GUID{eid: writerGUID.eid, prefix: writerGUID.prefix},
				}
				match = r
				println("adding reader due to heartbeat RX")
				defaultSession.addReader(r)
				break
			}
		}
	}

	if match != nil {
		if match.reliable && !final {
			// we have to send an ACKNACK now
			var set SeqNumSet
			if match.maxRxSeqNum >= hb.lastSeqNum {
				println("hb: acknack requested, up to date")
				set.bitmapBase = hb.firstSeqNum + 1
				set.numBits = 0
			} else {
				println("hb: acknack requested, acknack'ing multiple samples")
				set.bitmapBase = match.maxRxSeqNum + 1
				set.numBits = uint32(hb.lastSeqNum - match.maxRxSeqNum - 1)
				if set.numBits > 31 {
					set.numBits = 31
				}
				set.bitmap = []uint32{0xffffffff} // all bits acked for now
			}
			udpTxAckNack(r.srcGUIDPrefix, match.readerEID, match.writerGUID, set)
		} else {
			println("FINAL flag not set in heartbeat; not going to tx acknack")
		}
	} else {
		fmt.Printf("      couldn't find match for inbound heartbeat:\n")
		fmt.Printf("         %s => 0x%08x", writerGUID.prefix.String(), hb.readerEID)
	}
}
