// A master server keeps track of all public q2 servers.
// Q2 servers need to specifically be told to report
// to a master server.
package master

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/state"
)

const (
	DefaultListenPort    = 27900
	DefaultListenAddr    = "[::]" // IPv4/IPv6 all
	DefaultThinkInterval = 360    // secs
	DefaultPingInterval  = 360    // secs
	DefaultApiPort       = 3333
	DefaultApiAddr       = "[::]"
	DefaultApiEnabled    = false
)

// the all-knowning master server
type MasterServer struct {
	Address        string // IP or DNS name
	Port           int    // default 27900
	ApiEnabled     bool
	ApiIP          string
	ApiPort        int
	Clients        []MasterClient  // our known q2 server
	Conn           *net.PacketConn // the socket
	ThinkInterval  int             // seconds between thinks
	PingInterval   int
	ThinkFunc      func(m *MasterServer)
	ProcessFunc    func(m *MasterServer)
	ClientListFunc func(m *MasterServer, recip *net.Addr)
	HeartbeatFunc  func(m *MasterServer, from *net.Addr, info map[string]string)
	PingFunc       func(m *MasterServer, from *net.Addr) *MasterClient
	AckFunc        func(m *MasterServer, from *net.Addr)
	ShutdownFunc   func(m *MasterServer, from *net.Addr)
	Stats          MasterServerStats
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

// Setup a new server struct with default function calls and values
func NewMaster() *MasterServer {
	master := MasterServer{
		Address:    DefaultListenAddr,
		Port:       DefaultListenPort,
		ApiEnabled: DefaultApiEnabled,
		ApiIP:      DefaultApiAddr,
		ApiPort:    DefaultApiPort,
		Stats: MasterServerStats{
			StartTime: time.Now(),
		},
		ThinkInterval:  DefaultThinkInterval,
		ThinkFunc:      think,
		ClientListFunc: clientList,
		PingFunc:       ping,
		AckFunc:        ack,
		HeartbeatFunc:  heartbeat,
		ShutdownFunc:   shutdown,
		PingInterval:   DefaultPingInterval,
	}
	return &master
}

// start the actual server
func (m *MasterServer) Run() {
	log.Println("Starting up...")
	listenAddr := fmt.Sprintf("%s:%d", m.Address, m.Port)
	listener, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	m.Conn = &listener
	log.Println("Listening for Q2 Servers on", listenAddr)

	if m.ThinkFunc != nil {
		go m.ThinkFunc(m)
	}

	if m.ApiEnabled {
		go startAPIServer(m)
	}

	buf := make([]byte, 1024)
	for {
		count, addr, err := listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		go processMessage(m, &addr, buf[:count])
	}
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
			if msg.Index == len(msg.Data) {
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
func (m *MasterServer) MarshalClients() *message.Buffer {
	msg := message.Buffer{}
	for _, cl := range m.Clients {
		msg.Append(*cl.Marshal())
	}
	return &msg
}

// Write this MasterClient's IP and port in a format that can be sent
// as a response
func (cl *MasterClient) Marshal() *message.Buffer {
	msg := message.Buffer{}
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
	fmt.Printf("%s\n", hex.Dump(msg.Data))
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

// Periodically checks in on each client, pruning dead ones.
// Should be run concurrently
func think(m *MasterServer) {
	for {
		time.Sleep(time.Duration(m.ThinkInterval) * time.Second)
		for i := range m.Clients {
			if m.Clients[i].PendingAcks > 3 {
				removeClient(m, &m.Clients[i].Address)
				continue
			}
			needsPing := m.Clients[i].LastContact.Add(time.Duration(m.PingInterval) * time.Second)
			if time.Now().After(needsPing) {
				send("ping", m, &m.Clients[i].Address)
				m.Clients[i].PendingAcks++
				m.Clients[i].LastContact = time.Now()
			}
		}
	}
}

// Removes a client from the client slice.
// Returns the slice of clients removed
func removeClient(m *MasterServer, from *net.Addr) {
	oldClients := &m.Clients
	newClients := []MasterClient{}

	for i, cl := range *oldClients {
		if cl.Address.String() == (*from).String() {
			continue
		}
		newClients = append(newClients, (*oldClients)[i])
	}
	m.Clients = newClients
}

// for sending simple "ack"s and "ping"s
func send(cmd string, m *MasterServer, recip *net.Addr) {
	ack := message.Buffer{}
	ack.WriteLong(-1)
	ack.WriteData([]byte(cmd))
	(*m.Conn).WriteTo(ack.Data, *recip)
}

// Runs concurrently for every datagram recieved by the server
func processMessage(m *MasterServer, from *net.Addr, buf []byte) {
	msg := message.Buffer{
		Data: buf,
	}
	if msg.ReadLong() == -1 {
		tok := strings.Split(string(msg.ReadData(len(buf))), "\n")
		cmd := strings.Trim(tok[0], "\x00\x0a\x20\x09") // null, new line, space, tab

		switch cmd {
		case "getservers":
			if m.ClientListFunc != nil {
				m.ClientListFunc(m, from)
			}
		case "ping":
			if m.PingFunc != nil {
				m.PingFunc(m, from)
			}
		case "heartbeat":
			if m.HeartbeatFunc != nil {
				m.HeartbeatFunc(m, from, state.ParseInfoString(tok[1][1:]))
			}
		case "ack":
			if m.AckFunc != nil {
				m.AckFunc(m, from)
			}
		case "shutdown":
			if m.ShutdownFunc != nil {
				m.ShutdownFunc(m, from)
			}
		default:
			log.Println("Ignoring unknown command from", *from, cmd)
		}
	} else {
		msg.Rewind()
		if msg.ReadString() == "query\n" {
			if m.ClientListFunc != nil {
				m.ClientListFunc(m, from)
			}
		}
	}
}

// Someone requested a list of all Q2 servers we know about.
func clientList(m *MasterServer, recip *net.Addr) {
	m.Stats.GetServerHits++
	msg := message.Buffer{}
	msg.WriteLong(-1)
	msg.WriteData([]byte("servers ")) // note the space

	clients := m.MarshalClients()
	msg.Append(*clients)
	(*m.Conn).WriteTo(msg.Data, *recip)
	log.Println("sending client list to", *recip)
}

// Sent from client to us every 5-10ish or so minutes.
func heartbeat(m *MasterServer, from *net.Addr, info map[string]string) {
	cl := m.FindClient(*from)
	if cl == nil {
		cl = ping(m, from)
	}
	cl.Heartbeats++
	cl.LastContact = time.Now()
	cl.Hostname = info["hostname"]
	cl.GameDir = info["gamedir"]
	// game?
	mp, _ := strconv.Atoi(info["maxclients"])
	cl.MaxPlayers = mp
	send("ack", m, from)
	log.Println("heartbeat from", (*from).String(), "-", info["hostname"])
}

// An unfamiliar server started talking to us. Start tracking it.
func ping(m *MasterServer, from *net.Addr) *MasterClient {
	c := m.FindClient(*from)
	if c != nil {
		return c // we already have this one
	}

	tokens := strings.Split((*from).String(), ":")
	port, _ := strconv.Atoi(tokens[1])
	cl := MasterClient{
		Address: *from,
		IP:      net.ParseIP(tokens[0]),
		Port:    port,
	}
	m.Clients = append(m.Clients, cl)
	log.Println("adding client", (*from).String(), "-", len(m.Clients), "total")
	return &m.Clients[len(m.Clients)-1]
}

// A client sends us an ack when he "ping" them (from management)
func ack(m *MasterServer, from *net.Addr) {
	cl := m.FindClient(*from)
	if cl == nil {
		return
	}
	cl.Heartbeats++
	cl.LastContact = time.Now()
	cl.PendingAcks--
	sv := state.Server{Address: cl.IP.String(), Port: cl.Port}
	info, err := sv.FetchInfo()
	if err == nil {
		players := []MasterClientPlayer{}
		for _, p := range info.Players {
			players = append(players, MasterClientPlayer{
				Name:  p.Name,
				Score: p.Score,
				Ping:  p.Ping,
			})
		}
		cl.Players = players
		cl.Info = info.Server
	}
	log.Println("ack from", (*from).String())
}

// Clients issue shutdown msgs when they quit or go non-public
func shutdown(m *MasterServer, from *net.Addr) {
	cl := m.FindClient(*from)
	if cl == nil {
		return
	}
	removeClient(m, from)
	log.Println("shutdown issued from", (*from).String())
}
