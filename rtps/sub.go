package rtps

import (
	"fmt"
)

type dataCallback func(r *receiver, submsg *subMsg, scheme uint16, b []byte)
type msgCallback func(b []byte)

// Reader is a reader
type Reader struct {
	reliable    bool
	writerGUID  GUID
	readerEID   EntityID
	maxRxSeqNum SeqNum
	dataCB      dataCallback
	msgCB       msgCallback
}

type Sub struct {
	topicName string
	typeName  string
	readerEID EntityID
	dataCB    dataCallback
	msgCB     msgCallback
	reliable  bool
}

func AddUserSub(topicName, typeName string, cb msgCallback) {
	subEID := createUserID(ENTITYID_KIND_READER_NO_KEY)
	fmt.Printf("frudp_add_user_sub(%s, %s) on EID 0x%08x\r\n", topicName, typeName, subEID)
	defaultSession.addSub(&Sub{
		topicName: topicName,
		typeName:  typeName,
		readerEID: subEID,
		msgCB:     cb,
		reliable:  false,
	})
	//sedp_publish_sub(&sub); // can't do this yet; spdp hasn't started bcast
}
