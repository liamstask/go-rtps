package rtps

import (
	"bytes"
	"fmt"
)

const (
	UDPGuidPrefixLen  = 12
	EIDUnknown        = 0
	SPDPWriterID      = 0xc2000100
	SPDPReaderID      = 0xc7000100
	SEDPPubWriterID   = 0xc2030000
	SEDPPubReaderID   = 0xc7030000
	SEDPSubWriterID   = 0xc2040000
	SEDPSubReaderID   = 0xc7040000
	Magic             = 0x52545053 // RTPS in ASCII
	MY_RTPS_VENDOR_ID = 0x1234
)

const (
	NN_ENTITYID_UNKNOWN                                = 0x0
	NN_ENTITYID_PARTICIPANT                            = 0x1c1
	NN_ENTITYID_SEDP_BUILTIN_TOPIC_WRITER              = 0x2c2
	NN_ENTITYID_SEDP_BUILTIN_TOPIC_READER              = 0x2c7
	NN_ENTITYID_SEDP_BUILTIN_PUBLICATIONS_WRITER       = 0x3c2
	NN_ENTITYID_SEDP_BUILTIN_PUBLICATIONS_READER       = 0x3c7
	NN_ENTITYID_SEDP_BUILTIN_SUBSCRIPTIONS_WRITER      = 0x4c2
	NN_ENTITYID_SEDP_BUILTIN_SUBSCRIPTIONS_READER      = 0x4c7
	NN_ENTITYID_SPDP_BUILTIN_PARTICIPANT_WRITER        = 0x100c2
	NN_ENTITYID_SPDP_BUILTIN_PARTICIPANT_READER        = 0x100c7
	NN_ENTITYID_P2P_BUILTIN_PARTICIPANT_MESSAGE_WRITER = 0x200c2
	NN_ENTITYID_P2P_BUILTIN_PARTICIPANT_MESSAGE_READER = 0x200c7
	NN_ENTITYID_SOURCE_MASK                            = 0xc0
	NN_ENTITYID_SOURCE_USER                            = 0x00
	NN_ENTITYID_SOURCE_BUILTIN                         = 0xc0
	NN_ENTITYID_SOURCE_VENDOR                          = 0x40
	NN_ENTITYID_KIND_MASK                              = 0x3f
	NN_ENTITYID_KIND_WRITER_WITH_KEY                   = 0x02
	NN_ENTITYID_KIND_WRITER_NO_KEY                     = 0x03
	NN_ENTITYID_KIND_READER_NO_KEY                     = 0x04
	NN_ENTITYID_KIND_READER_WITH_KEY                   = 0x07
	NN_ENTITYID_ALLOCSTEP                              = 0x100
)

func vendorName(id VendorID) string {
	switch id {
	case 0x0101:
		return "RTI Connext"
	case 0x0102:
		return "PrismTech OpenSplice"
	case 0x0103:
		return "OCI OpenDDS"
	case 0x0104:
		return "MilSoft"
	case 0x0105:
		return "Gallium InterCOM"
	case 0x0106:
		return "TwinOaks CoreDX"
	case 0x0107:
		return "Lakota Technical Systems"
	case 0x0108:
		return "ICOUP Consulting"
	case 0x0109:
		return "ETRI"
	case 0x010a:
		return "RTI Connext Micro"
	case 0x010b:
		return "PrismTech Vortex Cafe"
	case 0x010c:
		return "PrismTech Vortex Gateway"
	case 0x010d:
		return "PrismTech Vortex Lite"
	case 0x010e:
		return "Technicolor Qeo"
	case 0x010f:
		return "eProsima"
	case 0x0120:
		return "PrismTech Vortex Cloud"
	case MY_RTPS_VENDOR_ID:
		return "go-rtps"
	default:
		return "unknown"
	}
}

// EntityID is an entity id
type EntityID uint32

func (eid EntityID) kind() uint8 {
	return uint8(eid & 0xff)
}

func (eid EntityID) key() []byte {
	return []byte{byte(eid>>8) & 0xff, byte(eid>>16) & 0xff, byte(eid>>24) & 0xff}
}

type VendorID uint16
type GUIDPrefix []byte

func newGUIDPrefix() GUIDPrefix {
	return make([]byte, UDPGuidPrefixLen)
}

func (gp GUIDPrefix) String() string {
	if gp == nil {
		return "<nil guid>"
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x%02x%02x-%02x%02x%02x%02x",
		gp[0], gp[1], gp[2], gp[3], gp[4], gp[5], gp[6], gp[7], gp[8], gp[9], gp[10], gp[11])
}

type GUID struct {
	prefix GUIDPrefix
	eid    EntityID
}

func (g *GUID) Equal(other *GUID) bool {
	return g.eid == other.eid && bytes.Equal(g.prefix, other.prefix)
}

// unknown Guid is all 0's
