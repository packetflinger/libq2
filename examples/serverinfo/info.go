package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/state"
)

func main() {
	srv := state.Server{
		Address: "tastyspleen.net",
		Port:    27916,
	}

	info, err := srv.FetchInfo()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(info.Server["version"])
}
