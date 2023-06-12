// demoprints simply reads a Quake 2 demo file and prints out any
// print messages found.
package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/demo"
	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/util"
)

func main() {
	// open the demo file
	demo, err := demo.OpenDM2File("../testdata/testduel.dm2")
	if err != nil {
		log.Println(err)
		return
	}

	// set a callback for parsing prints
	callback := message.MessageCallbacks{
		Print: func(p *message.Print) {
			fmt.Println(util.ConvertHighChars(p.String[:len(p.String)-1]))
		},
	}

	// parse all demo messages running our callback function
	// every time a print message is found
	demo.ParseDM2(callback)

	// clean up
	demo.Close()
}
