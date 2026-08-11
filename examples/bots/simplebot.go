package main

import (
	"flag"
	"log"

	"github.com/packetflinger/libq2/bot"
	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/player"

	pb "github.com/packetflinger/libq2/proto"
)

var (
	name   = flag.String("name", "totallynotabot", "The player name to use")
	server = flag.String("server", "frag.gr", "Q2 server ip/hostname to connect to")
	port   = flag.Int("port", 27910, "The port the server is listening on")
	debug  = flag.Bool("debug", false, "show way more information in the console")
)

func main() {
	flag.Parse()
	bot := bot.Bot{
		Version: "PFBot Test v1",
		Net: bot.Connection{
			Address: *server,
			Port:    *port,
		},
		User: player.Userinfo{
			"name": *name,
			"skin": "female/jezebel",
			"hand": "0",
			"rate": "15000",
		},
		CVars: map[string]string{
			"version":   "PFBot Test v2",
			"timescale": "1",
		},
		Debug: *debug,
	}

	bot.RegisterCallback(message.SVCPrint, printCallback)
	bot.RegisterCallback(message.SVCConfigString, configstringCallback)
	bot.RegisterCallback(message.SVCStuffText, stuffCallback)
	bot.RegisterCallback(message.SVCSpawnBaseline, baselineCallback)

	if err := bot.Run(); err != nil {
		log.Println(err)
	}
}

func printCallback(p any, _ *message.Buffer) {
	log.Print(p.(*pb.Print).GetData()) // newline already included
}

func configstringCallback(c any, _ *message.Buffer) {
	cs := c.(*pb.ConfigString)
	log.Printf("configstring - [%0.4d] %s\n", cs.GetIndex(), cs.GetData())
}

func stuffCallback(st any, _ *message.Buffer) {
	log.Printf("stuff - %q\n", st.(*pb.StuffText).GetData())
}

func baselineCallback(b any, _ *message.Buffer) {
	ent := b.(*pb.PackedEntity)
	log.Printf("baseline - %03d\n", ent.GetNumber())
}
