// An example Quake 2 master server
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/packetflinger/libq2/master"
)

var (
	listenPort = flag.Int("port", 27900, "Port to listen on")
	listenIP   = flag.String("addr", "[::]", "IP address to listen on")
)

func main() {
	flag.Parse()

	log.Printf("*** Quake 2 Master Server - (c) 2022-%d Packetflinger Industries ***\n", time.Now().Year())
	m := master.NewMaster()
	m.Address = *listenIP
	m.Port = *listenPort
	m.Refresh = true

	ctx, done := context.WithCancel(context.Background())
	defer done()

	m.Run(ctx)
}
