package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/bot"
	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/player"

	pb "github.com/packetflinger/libq2/proto"
)

func main() {
	bot := bot.Bot{
		Version: "PFBot Test v1",
		Net: bot.Connection{
			Address: "dev.frag.gr",
			Port:    27910,
		},
		User: player.Userinfo{
			Name: "test1",
			Skin: "female/jezebel",
			Hand: 0,
			Rate: 15000,
		},
		Debug: false,
	}

	bot.RegisterCallback(message.SVCPrint, printCallback)
	if err := bot.Run(); err != nil {
		log.Println(err)
	}
}

func printCallback(p any, _ *message.Buffer) {
	fmt.Println("Print callback:", p.(*pb.Print).GetData())
}
