// demoprints simply reads a Quake 2 demo file and prints out any
// print messages found.
package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/demo"
)

func main() {
	// open the demo file
	dm2, err := demo.NewDM2Demo("../../testdata/testduel.dm2")
	if err != nil {
		log.Println(err)
		return
	}

	err = dm2.Unmarshal()
	if err != nil {
		log.Println(err)
		return
	}

	for _, frame := range dm2.GetTextProto().GetFrames() {
		fmt.Println("frame!")
		for _, print := range frame.GetPrints() {
			fmt.Println(print.GetString_())
		}
	}
	// set a callback for parsing prints
	/*callback := message.MessageCallbacks{
		Print: func(p *message.Print) {
			fmt.Println(util.ConvertHighChars(p.String[:len(p.String)-1]))
		},
	}*/

	// parse all demo messages running our callback function
	// every time a print message is found
	//demo.ParseDM2(callback)
}
