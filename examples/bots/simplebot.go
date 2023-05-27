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
		ServerDataCB:   func(s m.ServerData) { fmt.Println("ServerData:", s) },
		ConfigStringCB: func(cs m.ConfigString) { fmt.Println("ConfigString:", cs) },
		BaselineCB:     func(b m.PackedEntity) { fmt.Println("Baseline:", b) },
		StuffCB:        func(s m.StuffText) { fmt.Println("Stuff:", s.String) },
		FrameCB:        func(f m.FrameMsg) { fmt.Printf("Frame [%d,%d]\n", f.Number, f.Delta) },
		PlayerStateCB:  func(p m.PackedPlayer) { fmt.Println("Playerstate") },
		EntityCB: func(ents []m.PackedEntity) {
			fmt.Printf("Entities [")
			for _, e := range ents {
				fmt.Printf("%d,", e.Number)
			}
			fmt.Printf("]\n")
		},
		PrintCB:       func(p m.Print) { fmt.Printf("Print: [%d] %s\n", p.Level, p.String) },
		LayoutCB:      func(l m.Layout) { fmt.Println("Layout") },
		CenterPrintCB: func(p m.CenterPrint) { fmt.Printf("CenterPrint: %s\n", p.Data) },
		SoundCB:       func(s m.PackedSound) { fmt.Println("Sound:", s) },
		TempEntCB:     func(t m.TemporaryEntity) { fmt.Println("TempEnt:", t) },
		Flash1CB:      func(f m.MuzzleFlash) { fmt.Println("MuzzleFlash1") },
		Flash2CB:      func(f m.MuzzleFlash) { fmt.Println("MuzzleFlash2") },
	}

	callbacks = m.MessageCallbacks{
		PrintCB: func(p m.Print) {
			fmt.Println(p.String)
		},
		CenterPrintCB: func(c m.CenterPrint) {
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
