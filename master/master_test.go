package master

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/packetflinger/libq2/message"
)

func TestFetchServers(t *testing.T) {
	oob := message.ConnectionlessPacket{
		Data: "getservers",
	}
	msg, err := oob.Send("master.quake.services", 27900)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("%s\n", hex.Dump(msg.Buffer))
	t.Error()
}
