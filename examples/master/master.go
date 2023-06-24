package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/packetflinger/libq2/master"
	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/state"
)

var (
	listenPort = flag.Int("port", 27900, "Port to listen on")
	listenIP   = flag.String("addr", "0.0.0.0", "IP address to listen on")
	pingEvery  = flag.Int("pingtime", 120, "Ping server if not heard from in this number of seconds")
	mgmtTime   = flag.Int("mgmttime", 31, "Check all servers every x seconds")
	foreground = flag.Bool("fg", false, "Log to stdout instead of file")
	logfile    = flag.String("logfile", "master.log", "The filename to use for the log")
	api        = flag.Bool("api", false, "Whether or not to enable the web API")
	apiPort    = flag.Int("apiport", 3333, "TCP port for web requests")
	apiIP      = flag.String("apiaddr", "0.0.0.0", "The IP address to listen on for web requests")
)

func main() {
	flag.Parse()
	if !*foreground {
		fp, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			fp.Close()
		}()
		log.SetOutput(io.Writer(fp))
	}

	log.Printf("*** Quake 2 Master Server - (c) 2022-%d Packetflinger Industries ***\n", time.Now().Year())
	log.Println("Starting...")

	master := master.MasterServer{
		Address: *listenIP,
		Port:    *listenPort,
		Stats: master.MasterServerStats{
			StartTime: time.Now(),
		},
	}
	listenAddr := fmt.Sprintf("%s:%d", master.Address, master.Port)
	listener, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	master.Conn = &listener
	log.Println("Listening for Q2 Servers on", listenAddr)

	go think(&master)
	if *api {
		go startAPIServer(&master)
	}

	for {
		buf := make([]byte, 1024)
		count, addr, err := listener.ReadFrom(buf)
		if err != nil {
			continue
		}
		go processMessage(&master, &addr, buf[:count])
	}
}

// periodically checks in on each client, pruning dead ones.
// Runs concurrently as a goroutine
func think(m *master.MasterServer) {
	for {
		time.Sleep(time.Duration(*mgmtTime) * time.Second)

		for i := range m.Clients {
			// dead client, ditch em
			if m.Clients[i].PendingAcks > 3 {
				removeClient(m, &m.Clients[i].Address)
				continue
			}
			needsPing := m.Clients[i].LastContact.Add(time.Duration(*pingEvery) * time.Second)
			if time.Now().After(needsPing) {
				send("ping", m, &m.Clients[i].Address)
				m.Clients[i].PendingAcks++
				m.Clients[i].LastContact = time.Now()
			}
		}
	}
}

// Runs concurrently for every datagram recieved by the server
func processMessage(master *master.MasterServer, from *net.Addr, buf []byte) {
	msg := message.MessageBuffer{
		Buffer: buf,
	}
	if msg.ReadLong() == -1 {
		tok := strings.Split(string(msg.ReadData(len(buf))), "\n")
		cmd := strings.Trim(tok[0], "\x00\x0a\x20\x09") // null, new line, space, tab

		switch cmd {
		case "getservers":
			sendClientList(master, from)
		case "ping":
			addClient(master, from)
		case "heartbeat":
			heartbeat(master, from, state.ParseInfoString(tok[1][1:]))
		case "ack":
			ack(master, from)
		case "shutdown":
			shutdown(master, from)
		default:
			log.Println("Ignoring unknown command from", *from, cmd)
		}
	} else {
		msg.Rewind()
		if msg.ReadString() == "query\n" {
			sendClientList(master, from)
		}
	}
}

// Someone requested a list of all Q2 servers we know about.
func sendClientList(m *master.MasterServer, recip *net.Addr) {
	m.Stats.GetServerHits++
	msg := message.MessageBuffer{}
	msg.WriteLong(-1)
	msg.WriteData([]byte("servers ")) // note the space

	clients := m.MarshalClients()
	msg.Append(*clients)
	(*m.Conn).WriteTo(msg.Buffer, *recip)
	log.Println("sending client list to", *recip)
}

// for sending simple "ack"s and "ping"s
func send(cmd string, m *master.MasterServer, recip *net.Addr) {
	ack := message.MessageBuffer{}
	ack.WriteLong(-1)
	ack.WriteData([]byte(cmd))
	(*m.Conn).WriteTo(ack.Buffer, *recip)
}

// Removes a client from the client slice.
// Returns the number of remaining clients
func removeClient(m *master.MasterServer, from *net.Addr) int {
	oldClients := &m.Clients
	newClients := []master.MasterClient{}

	for i, cl := range *oldClients {
		if cl.Address.String() == (*from).String() {
			continue
		}
		newClients = append(newClients, (*oldClients)[i])
	}
	m.Clients = newClients
	log.Println("removed client", (*from).String())
	return len(newClients)
}

// An unfamiliar server started talking to us. Start tracking it.
func addClient(m *master.MasterServer, from *net.Addr) *master.MasterClient {
	c := m.FindClient(*from)
	if c != nil {
		return c // we already have this one
	}

	tokens := strings.Split((*from).String(), ":")
	port, _ := strconv.Atoi(tokens[1])
	cl := master.MasterClient{
		Address: *from,
		IP:      net.ParseIP(tokens[0]),
		Port:    port,
	}
	m.Clients = append(m.Clients, cl)
	log.Println("adding client", (*from).String(), "-", len(m.Clients), "total")
	return &m.Clients[len(m.Clients)-1]
}

// Sent from client to us every 5-10ish or so minutes.
func heartbeat(m *master.MasterServer, from *net.Addr, info map[string]string) {
	cl := m.FindClient(*from)
	if cl == nil {
		cl = addClient(m, from)
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

// A client sends us an ack when he "ping" them (from management)
func ack(m *master.MasterServer, from *net.Addr) {
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
		players := []master.MasterClientPlayer{}
		for _, p := range info.Players {
			players = append(players, master.MasterClientPlayer{
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

// clients issue shutdown msgs when they quit or go non-public
func shutdown(m *master.MasterServer, from *net.Addr) {
	cl := m.FindClient(*from)
	if cl == nil {
		return
	}
	count := removeClient(m, from)
	log.Println("shutdown issued from", (*from).String(), "-", count, "total")
}

func startAPIServer(m *master.MasterServer) {
	type Server struct {
		Hostname   string
		Address    string
		IP         string
		Port       int
		Game       string
		Maxclients int
		Players    []master.MasterClientPlayer
		Info       map[string]string
	}
	apiHost := fmt.Sprintf("%s:%d", *apiIP, *apiPort)
	log.Printf("Listening for API requests on http://%s\n", apiHost)

	http.HandleFunc("/GetServers", func(w http.ResponseWriter, r *http.Request) {
		m.Stats.ApiHits++
		servers := []Server{}
		for _, s := range m.Clients {
			servers = append(servers, Server{
				Hostname:   s.Hostname,
				Address:    s.Address.String(),
				IP:         s.IP.String(),
				Port:       s.Port,
				Game:       s.GameDir,
				Maxclients: s.MaxPlayers,
				Players:    s.Players,
			})
		}

		js, err := json.MarshalIndent(servers, "", "  ")
		if err != nil {
			fmt.Fprintln(w, "500, internal server error")
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(js))
	})

	http.HandleFunc("/HealthCheck", func(w http.ResponseWriter, r *http.Request) {
		type Health struct {
			Uptime          string
			ServerCount     int
			TotalHeartbeats int
			APIRequests     int
			ServerRequests  int
			MemoryUsage     string
		}
		m.Stats.ApiHits++
		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)

		health := Health{
			Uptime:          time.Since(m.Stats.StartTime).String(),
			ServerCount:     len(m.Clients),
			TotalHeartbeats: m.HeartbeatCount(),
			APIRequests:     m.Stats.ApiHits,
			ServerRequests:  m.Stats.GetServerHits,
			MemoryUsage:     fmt.Sprintf("%d KiB", mem.Alloc/1024),
		}
		js, err := json.MarshalIndent(health, "", "  ")
		if err != nil {
			fmt.Fprintln(w, "500, internal server error")
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(js))
	})

	http.HandleFunc("/ServerInfo", func(w http.ResponseWriter, r *http.Request) {
		srv := r.URL.Query().Get("srv")
		if srv == "" {
			log.Println("GET /ServerInfo -", r.RemoteAddr, "- invalid srv arg")
			return
		}
		addr, err := net.ResolveUDPAddr("udp", srv)
		if err != nil {
			log.Println(err)
			return
		}
		s := m.FindClient(addr)
		server := Server{
			Hostname:   s.Hostname,
			Address:    s.Address.String(),
			IP:         s.IP.String(),
			Port:       s.Port,
			Game:       s.GameDir,
			Maxclients: s.MaxPlayers,
			Players:    s.Players,
			Info:       s.Info,
		}
		js, err := json.MarshalIndent(server, "", "  ")
		if err != nil {
			fmt.Fprintln(w, "500, internal server error")
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(js))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		m.Stats.ApiHits++
		fmt.Fprintf(w, "fuck off")
	})

	http.ListenAndServe(apiHost, nil)
}
