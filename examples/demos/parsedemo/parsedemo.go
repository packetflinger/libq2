package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/packetflinger/libq2/demo"
	"github.com/packetflinger/libq2/message"

	pb "github.com/packetflinger/libq2/proto"
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

	dm2.RegisterCallback(message.SVCPrint, func(a any) {
		pr := a.(*pb.Print)
		fmt.Printf("%s\n", pr.GetData())
	})

	err = dm2.Unmarshal()
	if err != nil {
		log.Fatalln(err)
	}
}
