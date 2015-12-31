package rtps

import ()

type Writer struct {
	readerGUID GUID
	writerEID  EntityID
}

type Pub struct {
	topicName        string
	typeName         string
	writerEID        EntityID
	maxTxSeqNumAvail SeqNum
	minTxSeqNumAvail SeqNum
	dataSubmsgs      []*submsgData // buffered data messages waiting to be sent
	nextSubmsgIdx    uint32
	nextSeqNum       SeqNum
	reliable         bool
}

func newPub(topicName, typeName string, writerID EntityID, dataSubmsgs []*submsgData) *Pub {
	if writerID == EIDUnknown {
		return nil // xxx: better err reporting
	}

	pub := &Pub{
		topicName:     topicName,
		typeName:      typeName,
		writerEID:     writerID,
		nextSubmsgIdx: 0,
		nextSeqNum:    1,
	}
	if dataSubmsgs != nil { // it's for a reliable connection
		pub.dataSubmsgs = dataSubmsgs
		pub.reliable = true
	}
	// else unreliable; fire-and-forget from the caller's memory

	defaultSession.pubs = append(defaultSession.pubs, pub)
	return pub
}

func (pub *Pub) rxAckNack(an *submsgAckNack, gp GUIDPrefix) {
	// see if we have any of the requested messages
	// for (int req_seq_num = acknack->reader_sn_state.bitmap_base.low;
	//      req_seq_num < acknack->reader_sn_state.bitmap_base.low +
	//                    acknack->reader_sn_state.num_bits + 1;
	//      req_seq_num++)
	// {
	//   //printf("     request for seq num %d\n", req_seq_num);
	//   for msg_idx := 0; msg_idx < pub->num_data_submsgs; msg_idx++ {
	//     frudp_submsg_data_t *data = pub->data_submsgs[msg_idx];
	//     //printf("       %d ?= %d\n", req_seq_num, data->writer_sn.low);
	//     if (data->writer_sn.low == req_seq_num)
	//     {
	//       //printf("        found it in the history cache! now i need to tx\n");
	//
	//       //printf("about to look up guid prefix: ");
	//       //print_guid_prefix(guid_prefix);
	//       //printf("\n");
	//
	//       // look up the GUID prefix of this reader, so we can call them back
	//       frudp_part_t *part = frudp_part_find(guid_prefix);
	//       if (!part) {
	//         printf("      woah there partner. you from around these parts?\n");
	//         return;
	//       }
	//
	//       frudp_msg_t *msg = frudp_init_msg((frudp_msg_t *)g_pub_tx_buf);
	//       fr_time_t t = fr_time_now();
	//       uint16_t submsg_wpos = 0;
	//
	//       frudp_submsg_t *ts_submsg = (frudp_submsg_t *)&msg->submsgs[submsg_wpos];
	//       ts_submsg->header.id = FRUDP_SUBMSG_ID_INFO_TS;
	//       ts_submsg->header.flags = FRUDP_FLAGS_LITTLE_ENDIAN;
	//       ts_submsg->header.len = 8;
	//       memcpy(ts_submsg->contents, &t, 8);
	//       submsg_wpos += 4 + 8;
	//       ///////////////////////////////////////////////////////////////////////
	//       //frudp_submsg_data_t *data_submsg = (frudp_submsg_data_t *)&msg->submsgs[submsg_wpos];
	//       memcpy(&msg->submsgs[submsg_wpos], data, 4 + data->header.len);
	//       //data_submsg, submsg, 4 + submsg->header.len);
	//       submsg_wpos += 4 + data->header.len;
	//
	//       frudp_submsg_heartbeat_t *hb_submsg = (frudp_submsg_heartbeat_t *)&msg->submsgs[submsg_wpos];
	//       hb_submsg->header.id = FRUDP_SUBMSG_ID_HEARTBEAT;
	//       hb_submsg->header.flags = FLAGS_SM_ENDIAN | FLAGS_ACKNACK_FINAL,
	//       hb_submsg->header.len = 28;
	//       hb_submsg->reader_id = data->reader_id;
	//       hb_submsg->writer_id = data->writer_id;
	//       //printf("hb writer id = 0x%08x\n", htonl(data->writer_id.u));
	//       hb_submsg->first_sn.low = 1; // todo
	//       hb_submsg->first_sn.high = 0; // todo
	//       hb_submsg->last_sn = data->writer_sn;
	//       static int hb_count = 0;
	//       hb_submsg->count = hb_count++;
	//
	//       submsg_wpos += 4 + hb_submsg->header.len;
	//
	//       int payload_len = &msg->submsgs[submsg_wpos] - ((uint8_t *)msg);
	//       //printf("         sending %d bytes\n", payload_len);
	//       frudp_tx(part->metatraffic_unicast_locator.addr.udp4.addr,
	//                part->metatraffic_unicast_locator.port,
	//                (const uint8_t *)msg, payload_len);
	//     }
	//   }
	// }
}
