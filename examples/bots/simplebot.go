package main

import (
	"fmt"
	"log"

	"github.com/packetflinger/libq2/bot"
	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/player"
)

func main() {
	// simple print for each message type
	callbacks := message.MessageCallbacks{
		ServerData:   func(s *message.ServerData) { fmt.Println("ServerData:", s) },
		ConfigString: func(cs *message.ConfigString) { fmt.Println("ConfigString:", cs) },
		Baseline:     func(b *message.PackedEntity) { fmt.Println("Baseline:", b) },
		Stuff:        func(s *message.StuffText) { fmt.Println("Stuff:", s.String) },
		Frame:        func(f *message.FrameMsg) { fmt.Printf("Frame [%d,%d]\n", f.Number, f.Delta) },
		PlayerState:  func(p *message.PackedPlayer) { fmt.Println("Playerstate") },
		Entity: func(ents []*message.PackedEntity) {
			fmt.Printf("Entities [")
			for _, e := range ents {
				fmt.Printf("%d,", e.Number)
			}
			fmt.Printf("]\n")
		},
		Print:       func(p *message.Print) { fmt.Printf("Print: [%d] %s\n", p.Level, p.String) },
		Layout:      func(l *message.Layout) { fmt.Println("Layout") },
		CenterPrint: func(p *message.CenterPrint) { fmt.Printf("CenterPrint: %s\n", p.Data) },
		Sound:       func(s *message.PackedSound) { fmt.Println("Sound:", s) },
		TempEnt:     func(t *message.TemporaryEntity) { fmt.Println("TempEnt:", t) },
		Flash1:      func(f *message.MuzzleFlash) { fmt.Println("MuzzleFlash1") },
		Flash2:      func(f *message.MuzzleFlash) { fmt.Println("MuzzleFlash2") },
	}

	// Just output print msgs
	_ = message.MessageCallbacks{
		Print: func(p *message.Print) {
			if p.Level == message.PrintLevelChat {
				fmt.Println(p.String)
			}
		},
	}

	_ = message.MessageCallbacks{
		ConfigString: func(cs *message.ConfigString) {
			fmt.Println("ConfigString:", cs)
			//fmt.Printf("%s", hex.Dump([]byte(cs.String)))
		},
		/*qLayout: func(l *message.Layout) {
			fmt.Println("Layout")
			fmt.Printf("%s", hex.Dump([]byte(l.Data)))
		},*/
	}

	bot := bot.Bot{
		Version: "PFBot Test v1",
		Net: bot.Connection{
			Address: "dev.frag.gr",
			Port:    27999,
		},
		User: player.Userinfo{
			Name: "test1",
			Skin: "female/jezebel",
			Hand: 0,
			Rate: 15000,
		},
		Debug: true,
	}

	if err := bot.Run(callbacks); err != nil {
		log.Println(err)
	}
}
