package master

import (
	"fmt"
	"net"

	"github.com/packetflinger/libq2/message"
)

type Client struct {
	Host string
	Port int
}

// FetchPublicServers will ask a master server for the list of servers sending
// heartbeats to it. The master server knows nothing more about these game
// servers than their IP and port and that they're alive and active.
func (m *MasterServer) FetchPublicServers() ([]MasterClient, error) {
	//var out []MasterClient
	connectionless := message.ConnectionlessPacket{
		Data: "getservers",
	}
	resp, err := connectionless.Send(m.Address, m.Port)
	if err != nil {
		return nil, fmt.Errorf("sending connectionless packet: %s", err)
	}
	resp.Seek(13) // swallow the header and "servers "
	return ParseMasterResponse(&resp), nil
}

// ParseMasterResponse will break up the bytes returned from the master server
// into a slice of MasterClient structs. The only the `IP` and `Port` fields
// will be populated.
//
// Currently only IPv4 responses are supported. Adding support for IPv6 will
// mostly likely require a protocol change so clients will know weather to read
// 4 or 16 bytes.
func ParseMasterResponse(data *message.Buffer) []MasterClient {
	var out []MasterClient
	for !data.AtEnd() {
		if data.Length-data.Index < 6 {
			break
		}
		out = append(out, MasterClient{
			IP:   net.ParseIP(fmt.Sprintf("%d.%d.%d.%d", data.ReadByte(), data.ReadByte(), data.ReadByte(), data.ReadByte())),
			Port: (data.ReadByte() << 8) + data.ReadByte(),
		})
	}
	return out
}
