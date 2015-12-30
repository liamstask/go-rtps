package rtps

import (
	"bytes"
	"fmt"
)

// session collects that that would otherwise be global
type session struct {
	readers           []*Reader
	writers           []*Writer
	discoParticipants []*Participant
	pubs              []*Pub
	subs              []*Sub
	spdp              SPDP
	sedp              SEDP
	// XXX: some kind of config for which transport(s) are operational? just UDP for now
}

var (
	defaultSession session
)

func (s *session) init() {
	s.spdp.init()
	s.sedp.init()

	if err := udpInit(); err != nil {
		fmt.Println("udp init err:", err.Error())
	}
}

func (s *session) start() {
	s.spdp.start()
	s.sedp.start()
}

func (s *session) addReader(r *Reader) {
	// XXX: locking
	for _, rdr := range s.readers {
		if rdr.writerGUID.Equal(&r.writerGUID) {
			return
		}
	}
	s.readers = append(s.readers, r)
}

func (s *session) addSub(sub *Sub) {
	// XXX: locking
	fmt.Printf("sub %d: 0x%x\n", len(s.subs), sub.readerEID) // eid printed with wrong endianness
	s.subs = append(s.subs, sub)
}

func (s *session) pubWithWriterID(id EntityID) *Pub {
	for _, pub := range s.pubs {
		if id == pub.writerEID {
			return pub
		}
	}
	return nil
}

func (s *session) findParticipant(gp GUIDPrefix) *Participant {
	for _, p := range s.discoParticipants {
		if bytes.Equal(gp, p.guidPrefix) {
			return p
		}
	}
	return nil
}
