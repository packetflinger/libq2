package main

import (
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/packetflinger/libq2/master"
)

var (
	listenPort = flag.Int("port", 27900, "Port to listen on")
	listenIP   = flag.String("addr", "[::]", "IP address to listen on")
	// pingEvery  = flag.Int("pingtime", 120, "Ping server if not heard from in this number of seconds")
	// mgmtTime   = flag.Int("mgmttime", 31, "Check all servers every x seconds")
	foreground = flag.Bool("fg", false, "Log to stdout instead of file")
	logfile    = flag.String("logfile", "master.log", "The filename to use for the log")
	api        = flag.Bool("api", false, "Whether or not to enable the web API")
	apiPort    = flag.Int("apiport", 3333, "TCP port for web requests")
	apiIP      = flag.String("apiaddr", "[::]", "The IP address to listen on for web requests")
)

func main() {
	flag.Parse()
	if !*foreground {
		fp, err := os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()
		log.SetOutput(io.Writer(fp))
	}

	log.Printf("*** Quake 2 Master Server - (c) 2022-%d Packetflinger Industries ***\n", time.Now().Year())
	m := master.NewMaster()
	m.Address = *listenIP
	m.Port = *listenPort
	m.ApiEnabled = *api
	m.ApiIP = *apiIP
	m.ApiPort = *apiPort
	m.Run()
}
