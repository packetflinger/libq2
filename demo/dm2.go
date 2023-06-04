package demo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/util"
)

type DM2File struct {
	Filename      string
	Handle        *os.File
	Position      int
	ParsingFrames bool
	Serverdata    message.ServerData
	Configstrings [message.MaxConfigStrings]message.ConfigString
	Baselines     [message.MaxEntities]message.PackedEntity
	Frames        []message.ServerFrame
	CurrentFrame  *message.ServerFrame
	PrevFrame     *message.ServerFrame
	LumpCallback  func([]byte)
	Callbacks     message.MessageCallbacks
}

func OpenDM2File(f string) (*DM2File, error) {
	if !util.FileExists(f) {
		return nil, errors.New("no such file")
	}

	fp, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	demo := DM2File{
		Filename: f,
		Handle:   fp,
	}

	return &demo, nil
}

func (demo *DM2File) Close() {
	if demo.Handle != nil {
		demo.Handle.Close()
	}
}

func (demo *DM2File) ParseDM2(extcb message.MessageCallbacks) error {
	intcb := demo.InternalCallbacks()
	for {
		lump, size, err := nextLump(demo.Handle, int64(demo.Position))
		if err != nil {
			return err
		}
		if size == 0 {
			break
		}
		demo.Position += size

		_, err = message.ParseMessageLump(message.NewMessageBuffer(lump), intcb, extcb)
		if err != nil {
			return err
		}

		// do what we need to with all the messages read
	}
	return nil
}

func nextLump(f *os.File, pos int64) ([]byte, int, error) {
	_, err := f.Seek(pos, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lumplen := make([]byte, 4)
	_, err = f.Read(lumplen)
	if err != nil {
		return []byte{}, 0, err
	}

	lenbuf := message.MessageBuffer{Buffer: lumplen, Index: 0}
	length := lenbuf.ReadLong()

	// EOF
	if length == -1 {
		return []byte{}, 0, nil
	}

	_, err = f.Seek(pos+4, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lump := make([]byte, length)
	read, err := f.Read(lump)
	if err != nil {
		return []byte{}, 0, err
	}

	return lump, read + 4, nil
}

// Turn a parsed demo structure back into a binary file
func (demo *DM2File) Write() {
	msg := message.MessageBuffer{}

	//fmt.Println(demo.Serverdata)
	// first do serverdata
	msg.WriteByte(message.SVCServerData)
	msg.Append(demo.Serverdata.Marshal())

	// then configstrings
	for _, cs := range demo.Configstrings {
		msg.WriteByte(message.SVCConfigString)
		msg.Append(cs.Marshal())
	}

	/*
		// then baselines
		baselines := m.MessageBuffer{}
		nilEntity := m.PackedEntity{}
		for _, bl := range demo.Baselines {
			baselines.WriteDeltaEntity(bl, nilEntity)
		}
		msg.Append(&baselines)
	*/
	fmt.Printf("%s\n", hex.Dump(msg.Buffer))
}

func (demo *DM2File) MarshalLump() {

}

// Setup the callbacks for demo parsing. Stores data in the appropriate
// spots as it's parsed for later use
func (demo *DM2File) InternalCallbacks() message.MessageCallbacks {
	cb := message.MessageCallbacks{
		ServerData: func(s *message.ServerData) {
			demo.Serverdata = *s
		},
		ConfigString: func(cs *message.ConfigString) {
			if !demo.ParsingFrames {
				demo.Configstrings[cs.Index] = *cs
			} else {
				demo.CurrentFrame.Strings = append(demo.CurrentFrame.Strings, *cs)
			}
		},
		Baseline: func(b *message.PackedEntity) {
			demo.Baselines[b.Number] = *b
		},
		Stuff: func(s *message.StuffText) {
			if demo.ParsingFrames {
				demo.CurrentFrame.Stuffs = append(demo.CurrentFrame.Stuffs, *s)
			}
			if s.String == "precache\n" {
				demo.ParsingFrames = true
			}
		},
		Frame: func(f *message.FrameMsg) {
			demo.PrevFrame = demo.CurrentFrame
			demo.CurrentFrame = &message.ServerFrame{}
			demo.CurrentFrame.Frame = *f
		},
		PlayerState: func(p *message.PackedPlayer) {
			demo.CurrentFrame.Playerstate = *p
		},
		Entity: func(ents []*message.PackedEntity) {
			for _, e := range ents {
				demo.CurrentFrame.Entities[e.Number] = *e
			}
		},
		Print: func(p *message.Print) {
			demo.CurrentFrame.Prints = append(demo.CurrentFrame.Prints, *p)
		},
		Layout: func(l *message.Layout) {
			demo.CurrentFrame.Layouts = append(demo.CurrentFrame.Layouts, *l)
		},
		CenterPrint: func(p *message.CenterPrint) {
			demo.CurrentFrame.Centerprinters = append(demo.CurrentFrame.Centerprinters, *p)
		},
		Sound: func(s *message.PackedSound) {
			demo.CurrentFrame.Sounds = append(demo.CurrentFrame.Sounds, *s)
		},
		TempEnt: func(t *message.TemporaryEntity) {
			demo.CurrentFrame.TempEntities = append(demo.CurrentFrame.TempEntities, *t)
		},
		Flash1: func(f *message.MuzzleFlash) {
			demo.CurrentFrame.Flash1 = append(demo.CurrentFrame.Flash1, *f)
		},
		Flash2: func(f *message.MuzzleFlash) {
			demo.CurrentFrame.Flash2 = append(demo.CurrentFrame.Flash2, *f)
		},
	}
	return cb
}
