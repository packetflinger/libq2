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
		for _, print := range frame.GetPrints() {
			fmt.Println(print.GetData())
		}
	}
}
