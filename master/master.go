// A master server keeps track of all public q2 servers.
// Q2 servers need to specifically be told to report
// to a master server.
package master

import (
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/packetflinger/libq2/message"
)

// the all-knowning master server
type MasterServer struct {
	Address string          // IP or DNS name
	Port    int             // default 27900
	Clients []MasterClient  // our known q2 server
	Conn    *net.PacketConn // the socket
	Stats   MasterServerStats
}

type MasterServerStats struct {
	StartTime     time.Time // when the server started
	ApiHits       int       // how many times the API has been queried
	GetServerHits int       // how many times GetServers/query was issued
}

// A public Q2 server, also a client for the master
type MasterClient struct {
	Address     net.Addr // ip and port (192.0.2.1:27910)
	IP          net.IP
	Port        int
	Hostname    string
	GameDir     string
	MaxPlayers  int
	Players     []MasterClientPlayer
	LastContact time.Time
	Heartbeats  int
	Missedbeats int
	PendingAcks int
	Active      bool
	Info        map[string]string
}

type MasterClientPlayer struct {
	Name  string // 15 chars max
	Score int    // specs will be 0
	Ping  int
}

// Grab all servers
func (m MasterServer) FetchServers() ([]MasterClient, error) {
	clients := []MasterClient{}
	req := message.ConnectionlessPacket{
		Data: "getservers",
	}
	msg, err := req.Send(m.Address, m.Port)
	if err != nil {
		return clients, err
	}

	msg.ReadLong() // eat the sequence
	if string(msg.ReadData(7)) == "servers" {
		for {
			if msg.Index == len(msg.Buffer) {
				break
			}
			clients = append(clients, MasterClient{
				IP:   msg.ReadData(4),
				Port: int(msg.ReadShort()),
			})
		}
	}

	return clients, nil
}

// Write all MasterClient's info to a buffer for responding
func (m *MasterServer) MarshalClients() *message.MessageBuffer {
	msg := message.MessageBuffer{}
	for _, cl := range m.Clients {
		msg.Append(*cl.Marshal())
	}
	return &msg
}

// Write this MasterClient's IP and port in a format that can be sent
// as a response
func (cl *MasterClient) Marshal() *message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteData([]byte(cl.IP))

	// reversed byte-order from msg.WriteShort()
	port := []byte{
		byte((cl.Port >> 8) & 0xff),
		byte(cl.Port & 0xff),
	}
	msg.WriteData(port)
	return &msg
}

// dont do this
func (m MasterServer) SendClientList() {
	msg := m.MarshalClients()
	fmt.Printf("%s\n", hex.Dump(msg.Buffer))
}

func (m *MasterServer) FindClient(cl net.Addr) *MasterClient {
	for i, c := range m.Clients {
		if c.Address.String() == cl.String() {
			return &m.Clients[i]
		}
	}
	return nil
}

func (m *MasterServer) HeartbeatCount() int {
	total := 0
	for _, cl := range m.Clients {
		total += cl.Heartbeats
	}
	return total
}
