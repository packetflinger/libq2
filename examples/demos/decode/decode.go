// decode simply reads a Quake 2 demo file and converts it from binary to
// a text proto format.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/packetflinger/libq2/demo"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	demofile = flag.String("demofile", "", "path to .dm2 file")
)

func main() {
	flag.Parse()

	dm2, err := demo.NewDM2Demo(*demofile)
	if err != nil {
		log.Fatalln(err)
	}

	err = dm2.Unmarshal()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(prototext.Format(dm2.GetTextProto()))
}
