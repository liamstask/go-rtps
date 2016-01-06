package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/liamstask/go-rtps/rtps"
)

func main() {
	rtps.Setup()
	rtps.AddUserSub("chatter", "std_msgs::msg::dds_::String_", func(b []byte) {
		sz := binary.LittleEndian.Uint32(b[0:])
		msg := string(b[4 : 4+sz])
		fmt.Printf("I heard: %s\n", msg)
	})
	rtps.Start()

	for {
		time.Sleep(time.Second * 10)
	}
}
