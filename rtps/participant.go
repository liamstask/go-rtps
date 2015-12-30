package rtps

import (
	"time"
)

// "allows a participant to indicate that it only contains a
// subset of the possible builtin endpoints"
// bitmask of _BUILTIN_ENDPOINT_ values below
type builtinEndpointSet uint32

const (
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_ANNOUNCER       = (1 << 0)
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_DETECTOR        = (1 << 1)
	NN_DISC_BUILTIN_ENDPOINT_PUBLICATION_ANNOUNCER       = (1 << 2)
	NN_DISC_BUILTIN_ENDPOINT_PUBLICATION_DETECTOR        = (1 << 3)
	NN_DISC_BUILTIN_ENDPOINT_SUBSCRIPTION_ANNOUNCER      = (1 << 4)
	NN_DISC_BUILTIN_ENDPOINT_SUBSCRIPTION_DETECTOR       = (1 << 5)
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_PROXY_ANNOUNCER = (1 << 6) // undefined meaning
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_PROXY_DETECTOR  = (1 << 7) // undefined meaning
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_STATE_ANNOUNCER = (1 << 8) // undefined meaning
	NN_DISC_BUILTIN_ENDPOINT_PARTICIPANT_STATE_DETECTOR  = (1 << 9) // undefined meaning
	NN_BUILTIN_ENDPOINT_PARTICIPANT_MESSAGE_DATA_WRITER  = (1 << 10)
	NN_BUILTIN_ENDPOINT_PARTICIPANT_MESSAGE_DATA_READER  = (1 << 11)
)

type Participant struct {
	protoVer         ProtoVersion
	vid              VendorID
	guidPrefix       GUIDPrefix
	expectsInlineQoS bool
	defaultUcastLoc  locator
	defaultMcastLoc  locator
	metaUcastLoc     locator
	metaMcastLoc     locator
	leaseDuration    time.Duration
	builtinEndpoints builtinEndpointSet
}
