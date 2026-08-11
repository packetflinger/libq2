// decode simply reads a Quake 2 demo file and converts it from binary to
// a text proto format.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/packetflinger/libq2/demo"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	demofile = flag.String("demofile", "", "path to .dm2 file")
)

func main() {
	flag.Parse()
	content, err := os.ReadFile(*demofile)
	if err != nil {
		log.Fatalln(err)
	}
	dm2 := demo.NewDM2Parser()
	err = dm2.Unmarshal(content)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(prototext.Format(dm2.GetTextProto()))
}
