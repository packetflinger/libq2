// A master is an index of public Q2 servers. Servers need to be instructed to
// report to a specific master. All versions of R1Q2 and Q2Pro have
// `master.q2servers.com` hard coded as the default master. Servers will not
// report to a master unless they are set as public: `set public 1`
//
// A server will send a heartbeat every few minutes to announce it's still
// alive.
//
// Server commands:
//
//	`listmasters` - Show the currently configured masters.
//	`setmaster`   - Sets a space-delimited list of masters to use. To remove
//	                a master, issue this command again minus the server to
//	                remove.
//
// Servers can typically support up to 8 different masters.
package master

import (
	"context"
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
	DefaultListenAddr    = "[::]" // Any IPv4/IPv6
	DefaultThinkInterval = 360    // secs
	DefaultPingInterval  = 360    // secs
)

// the all-knowning master server
type MasterServer struct {
	Address       string            // IP or DNS name
	Clients       []MasterClient    // our known q2 server
	Conn          *net.PacketConn   // the socket
	GeoIPs        *state.GeoIPList  // the ip-country list
	PingInterval  int               // seconds
	Port          int               // default 27900
	Refresh       bool              // fetch player info
	Stats         MasterServerStats // stats for this master
	ThinkInterval int               // seconds
	Verbose       bool              // be extra mouthy

	AckFunc        func(m *MasterServer, from *net.Addr)
	ClientListFunc func(m *MasterServer, recip *net.Addr)
	HeartbeatFunc  func(m *MasterServer, from *net.Addr, info map[string]string)
	PingFunc       func(m *MasterServer, from *net.Addr) *MasterClient
	ProcessFunc    func(m *MasterServer)
	ShutdownFunc   func(m *MasterServer, from *net.Addr)
	ThinkFunc      func(ctx context.Context, m *MasterServer)
}

type MasterServerStats struct {
	GetServerHits int       // how many times GetServers/query was issued
	PlayerCount   int       // how many players are known?
	ServerCount   int       // how many q2 servers are registered
	StartTime     time.Time // when the server started
}

// A public Q2 server, also a client for the master
type MasterClient struct {
	Active       bool
	Address      net.Addr // ip and port (192.0.2.1:27910)
	Country      string   // 2 letter code, e.g. "US", "AU", "IE"
	CurrentMap   string
	FirstContact time.Time
	GameDir      string
	Heartbeats   int
	Hostname     string
	Info         map[string]string
	IP           net.IP
	LastContact  time.Time
	MaxPlayers   int
	Missedbeats  int
	Passworded   bool
	PendingAcks  int
	Players      []MasterClientPlayer
	Port         int
	Software     string
}

type MasterClientPlayer struct {
	ConnectTime time.Time // first seen
	Country     string    // 2 char code
	Name        string    // 15 chars max
	Ping        int       // in milliseconds
	Score       int       // specs will be 0
}

// Setup a new server struct with default function calls and values
func NewMaster() *MasterServer {
	master := MasterServer{
		Address: DefaultListenAddr,
		Port:    DefaultListenPort,
		Stats: MasterServerStats{
			StartTime: time.Now(),
		},
		ThinkInterval:  DefaultThinkInterval,
		ThinkFunc:      Think,
		ClientListFunc: ClientList,
		PingFunc:       Ping,
		AckFunc:        Ack,
		HeartbeatFunc:  Heartbeat,
		ShutdownFunc:   Shutdown,
		PingInterval:   DefaultPingInterval,
		Refresh:        false,
		Verbose:        false,
	}
	return &master
}

// start the actual server
func (m *MasterServer) Run(ctx context.Context) {
	listenAddr := fmt.Sprintf("%s:%d", m.Address, m.Port)
	listener, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	m.Conn = &listener
	log.Println("Listening for Q2 Servers on", listenAddr)
	if m.ThinkFunc != nil {
		go m.ThinkFunc(ctx, m)
	}
	if m.Refresh {
		go m.DetailRefresher(ctx)
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
	var clients []MasterClient
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
	msg := message.NewEmptyBuffer()
	for _, cl := range m.Clients {
		msg.Append(*cl.Marshal())
	}
	return &msg
}

// Write this MasterClient's IP and port in a format that can be sent as a
// response
func (cl *MasterClient) Marshal() *message.Buffer {
	msg := message.NewEmptyBuffer()
	msg.WriteData([]byte(cl.IP))

	// reversed byte-order from msg.WriteShort()
	port := []byte{
		byte((cl.Port >> 8) & 0xff),
		byte(cl.Port & 0xff),
	}
	msg.WriteData(port)
	return &msg
}

// Get a pointer to the client struct related to this IP address
func (m *MasterServer) FindClient(cl net.Addr) *MasterClient {
	for i, c := range m.Clients {
		if c.Address.String() == cl.String() {
			return &m.Clients[i]
		}
	}
	return nil
}

// Calculate the total number of heartbeats this server has seen from all
// clients.
func (m *MasterServer) HeartbeatCount() int {
	total := 0
	for _, cl := range m.Clients {
		total += cl.Heartbeats
	}
	return total
}

// Periodically checks in on each client, pruning dead ones. This should be run
// concurrently.
func Think(ctx context.Context, m *MasterServer) {
	ticker := time.NewTicker(time.Duration(m.ThinkInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if m.Verbose {
				log.Println("ending Think thread")
			}
			return
		case <-ticker.C:
			for i := range m.Clients {
				if m.Clients[i].PendingAcks > 3 {
					RemoveClient(m, &m.Clients[i].Address)
					continue
				}
				needsPing := m.Clients[i].LastContact.Add(time.Duration(m.PingInterval) * time.Second)
				if time.Now().After(needsPing) {
					Send("ping", m, &m.Clients[i].Address)
					m.Clients[i].PendingAcks++
					m.Clients[i].LastContact = time.Now()
				}
			}
		}
	}
}

// Removes a client from the client slice.
func RemoveClient(m *MasterServer, from *net.Addr) {
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

// For sending simple "ack"s and "ping"s
func Send(cmd string, m *MasterServer, recip *net.Addr) {
	ack := message.NewEmptyBuffer()
	ack.WriteLong(-1)
	ack.WriteData([]byte(cmd))
	(*m.Conn).WriteTo(ack.Data, *recip)
}

// Runs concurrently for every datagram recieved by the master
func processMessage(m *MasterServer, from *net.Addr, buf []byte) {
	msg := message.NewBuffer(buf)
	if msg.ReadLong() == -1 {
		tok := strings.Split(string(msg.ReadData(msg.UnreadSize())), "\n")
		if len(tok) == 0 {
			return
		}
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
				if len(tok) < 2 || tok[1] == "" {
					log.Printf("invalid heartbeat format from %q, ignoring: %v", (*from).String(), tok)
					return
				}
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
			log.Printf("Ignoring unknown command %q from %s\n", cmd, (*from).String())
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
func ClientList(m *MasterServer, recip *net.Addr) {
	m.Stats.GetServerHits++
	msg := message.NewEmptyBuffer()
	msg.WriteLong(-1)
	msg.WriteData([]byte("servers ")) // note the space
	clients := m.MarshalClients()
	msg.Append(*clients)
	(*m.Conn).WriteTo(msg.Data, *recip)
}

// Sent from client to us every 5-10ish or so minutes.
func Heartbeat(m *MasterServer, from *net.Addr, info map[string]string) {
	cl := m.FindClient(*from)
	if cl == nil {
		cl = Ping(m, from)
	}
	cl.Heartbeats++
	cl.LastContact = time.Now()
	cl.Hostname = info["hostname"]
	cl.GameDir = info["gamename"]
	cl.CurrentMap = info["mapname"]
	cl.Passworded = info["needpass"] == "1"
	cl.Software = info["version"]
	mp, err := strconv.Atoi(info["maxclients"])
	if err != nil {
		log.Printf("invalid maxclients value %q defaulting to 8\n", info["maxclients"])
		mp = 8
	}
	cl.MaxPlayers = mp
	if m.GeoIPs != nil {
		ip, _, err := net.SplitHostPort((*from).String())
		if err != nil {
			log.Printf("unable to split %q into host/port for location: %v\n", (*from).String(), err)
		} else {
			cl.Country = m.GeoIPs.Lookup(ip)
		}
	}
	Send("ack", m, from)
	if m.Verbose {
		log.Printf("heartbeat from %s - %s\n", (*from).String(), info["hostname"])
	}
}

// An unfamiliar server started talking to us. Start tracking it.
func Ping(m *MasterServer, from *net.Addr) *MasterClient {
	c := m.FindClient(*from)
	if c != nil {
		return c // we already have this one
	}
	tokens := strings.Split((*from).String(), ":")
	if len(tokens) != 2 {
		log.Printf("malformed addr %q, ignoring ping\n", (*from).String())
		return nil
	}
	port, err := strconv.Atoi(tokens[1])
	if err != nil {
		log.Printf("ping - unable to parse port %q, defaulting to 27900\n", tokens[1])
	}
	cl := MasterClient{
		Address:      *from,
		IP:           net.ParseIP(tokens[0]),
		Port:         port,
		FirstContact: time.Now(),
	}
	m.Clients = append(m.Clients, cl)
	log.Println("adding client", (*from).String(), "-", len(m.Clients), "total")
	return &m.Clients[len(m.Clients)-1]
}

// A client sends us an Ack when we "ping" them (from management)
func Ack(m *MasterServer, from *net.Addr) {
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
	if m.Verbose {
		log.Println("ack from", (*from).String())
	}
}

// Clients issue Shutdown msgs when they quit or go non-public
func Shutdown(m *MasterServer, from *net.Addr) {
	log.Println("shutdown issued from", (*from).String())
	cl := m.FindClient(*from)
	if cl == nil {
		return
	}
	RemoveClient(m, from)
}

// Once a minute, issue an out-of-band "info" request to each server to fetch
// the map and status for of the players. This should be run as a concurrent
// goroutine.
func (m *MasterServer) DetailRefresher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if m.Verbose {
				log.Println("context finished, stopping DetailRefresher thread")
			}
			return
		case <-ticker.C:
			playercount := 0
			servercount := 0
			for i, s := range m.Clients {
				srv := state.Server{Address: s.IP.String(), Port: s.Port}
				info, err := srv.FetchInfo()
				if err != nil {
					log.Printf("error fetching info for %q: %v", s.Hostname, err)
					continue
				}
				m.Clients[i].CurrentMap = info.Server["mapname"]
				m.Clients[i].Hostname = info.Server["hostname"]
				m.Clients[i].GameDir = info.Server["gamename"]
				mp, err := strconv.Atoi(info.Server["maxclients"])
				if err != nil {
					mp = 8
				}
				m.Clients[i].MaxPlayers = mp

				var pls []MasterClientPlayer
				if len(info.Players) > 0 {
					for _, p := range info.Players {
						when := time.Now()
						for _, x := range m.Clients[i].Players {
							if x.Name == p.Name {
								when = x.ConnectTime
							}
						}
						pls = append(pls, MasterClientPlayer{
							Name:        p.Name,
							Score:       p.Score,
							Ping:        p.Ping,
							ConnectTime: when,
						})
						playercount++
					}
				}
				m.Clients[i].Players = pls
				m.Stats.PlayerCount = playercount
				servercount++
			}
			m.Stats.ServerCount = servercount
		}
	}
}
