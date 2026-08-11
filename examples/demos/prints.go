// demoprints simply reads a Quake 2 demo file and prints out any
// print messages found.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/packetflinger/libq2/demo"
	"github.com/packetflinger/libq2/util"
)

func main() {
	content, err := os.ReadFile("../../testdata/testduel.dm2")
	if err != nil {
		log.Fatalln(err)
	}

	dm2 := demo.NewDM2Parser()
	err = dm2.Unmarshal(content)
	if err != nil {
		log.Println(err)
		return
	}

	for _, frame := range dm2.GetTextProto().GetFrames() {
		for _, print := range frame.GetPrints() {
			// newlines are included in the messages
			fmt.Print(util.ConvertHighChars(print.GetData()))
		}
	}
}
