package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/master"
)

func main() {
	srv := master.MasterServer{
		//Address: "master.quetoo.org",
		//Port:    1996,
		Address: "master.q2servers.com",
		Port:    27900,
	}

	clients, err := srv.FetchServers()
	if err != nil {
		log.Fatalln(err)
	}
	for _, cl := range clients {
		fmt.Printf("%s:%d\n", cl.IP, cl.Port)
	}
}
