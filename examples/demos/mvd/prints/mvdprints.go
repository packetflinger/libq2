package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/demo"

	pb "github.com/packetflinger/libq2/proto"
)

func main() {
	testdemo := "../../../../testdata/test.mvd2"
	parser, err := demo.NewMVD2Parser(testdemo)
	if err != nil {
		log.Fatalln(err)
	}

	parser.RegisterCallback(demo.MVDSvcPrint, func(data any) {
		print := data.(*pb.Print)
		fmt.Print(print.Data) // has newlines built-in
	})

	_, err = parser.Unmarshal()
	if err != nil {
		log.Fatalln(err)
	}
}
