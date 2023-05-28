package main

import (
	"fmt"
	"log"

	b "github.com/packetflinger/libq2/bot"
	m "github.com/packetflinger/libq2/message"
	p "github.com/packetflinger/libq2/player"
)

func main() {
	// simple print for each message type
	callbacks := m.MessageCallbacks{
		ServerData:   func(s *m.ServerData) { fmt.Println("ServerData:", s) },
		ConfigString: func(cs *m.ConfigString) { fmt.Println("ConfigString:", cs) },
		Baseline:     func(b *m.PackedEntity) { fmt.Println("Baseline:", b) },
		Stuff:        func(s *m.StuffText) { fmt.Println("Stuff:", s.String) },
		Frame:        func(f *m.FrameMsg) { fmt.Printf("Frame [%d,%d]\n", f.Number, f.Delta) },
		PlayerState:  func(p *m.PackedPlayer) { fmt.Println("Playerstate") },
		Entity: func(ents []*m.PackedEntity) {
			fmt.Printf("Entities [")
			for _, e := range ents {
				fmt.Printf("%d,", e.Number)
			}
			fmt.Printf("]\n")
		},
		Print:       func(p *m.Print) { fmt.Printf("Print: [%d] %s\n", p.Level, p.String) },
		Layout:      func(l *m.Layout) { fmt.Println("Layout") },
		CenterPrint: func(p *m.CenterPrint) { fmt.Printf("CenterPrint: %s\n", p.Data) },
		Sound:       func(s *m.PackedSound) { fmt.Println("Sound:", s) },
		TempEnt:     func(t *m.TemporaryEntity) { fmt.Println("TempEnt:", t) },
		Flash1:      func(f *m.MuzzleFlash) { fmt.Println("MuzzleFlash1") },
		Flash2:      func(f *m.MuzzleFlash) { fmt.Println("MuzzleFlash2") },
	}

	// Just output print msgs
	_ = m.MessageCallbacks{
		Print: func(p *m.Print) {
			fmt.Println(p.String)
		},
		CenterPrint: func(c *m.CenterPrint) {
			fmt.Println(c.Data)
		},
	}

	bot := b.Bot{
		Version: "PFBot Test v1",
		Net: b.Connection{
			Address: "frag.gr",
			Port:    27910,
		},
		User: p.Userinfo{
			Name: "totallynotabot",
			Skin: "female/jezebel",
			Hand: 0,
			Rate: 1000,
		},
		Debug: false,
	}

	if err := bot.Run(callbacks); err != nil {
		log.Println(err)
	}
}
