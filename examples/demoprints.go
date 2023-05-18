// demoprints simply reads a Quake 2 demo file and prints out any
// print messages found.
package main

import (
	"log"

	d "github.com/packetflinger/libq2/demo"
	m "github.com/packetflinger/libq2/message"
)

func main() {
	// open the demo file
	demo, err := d.OpenDM2File("../testdata/test.dm2")
	if err != nil {
		log.Println(err)
		return
	}

	// set a callback for parsing prints
	callback := m.MessageCallbacks{
		PrintCB: func(p m.Print) {
			log.Println(p.String)
		},
	}

	// parse all demo messages running our callback function
	// every time a print message is found
	demo.ParseDM2(callback)

	// clean up
	demo.Close()
}
