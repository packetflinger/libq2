package state

import (
	"fmt"
	"net"
	"strconv"
)

type Server struct {
	Address  string // IP or DNS name
	Port     int
	Password string // rcon
}

func NewServer(srv string) (Server, error) {
	if srv == "" {
		return Server{}, fmt.Errorf("NewServer(): empty input")
	}
	address, p, err := net.SplitHostPort(srv)
	if err != nil {
		return Server{}, fmt.Errorf("error splitting host and port: %v", err)
	}
	port, err := strconv.ParseInt(p, 10, 16)
	if err != nil {
		return Server{}, fmt.Errorf("NewServer(%q): invalid port, number 1-65535 required", srv)
	}
	return Server{Address: address, Port: int(port)}, nil
}
