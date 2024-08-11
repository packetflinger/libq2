package master

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

func startAPIServer(m *MasterServer) {
	type Server struct {
		Hostname   string
		Address    string
		IP         string
		Port       int
		Game       string
		Maxclients int
		Players    []MasterClientPlayer
		Info       map[string]string
	}
	apiHost := fmt.Sprintf("%s:%d", m.ApiIP, m.ApiPort)
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
