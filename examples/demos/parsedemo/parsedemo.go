package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/packetflinger/libq2/demo"
	"github.com/packetflinger/libq2/message"

	pb "github.com/packetflinger/libq2/proto"
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
	dm2.RegisterCallback(message.SVCPrint, func(a any) {
		pr := a.(*pb.Print)
		fmt.Printf("%s", pr.GetData()) // the \n is already on each line
	})
	err = dm2.Unmarshal(content)
	if err != nil {
		log.Fatalln(err)
	}
}
