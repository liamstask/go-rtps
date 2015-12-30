package rtps

import ()

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
