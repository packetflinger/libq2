package demo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	m "github.com/packetflinger/libq2/message"
	u "github.com/packetflinger/libq2/util"
)

type DM2File struct {
	Filename      string
	Handle        *os.File
	Position      int
	ParsingFrames bool
	Serverdata    m.ServerData
	Configstrings [m.MaxConfigStrings]m.ConfigString
	Baselines     [m.MaxEntities]m.PackedEntity
	Frames        []m.ServerFrame
	CurrentFrame  *m.ServerFrame
	PrevFrame     *m.ServerFrame
	LumpCallback  func([]byte)
	Callbacks     m.MessageCallbacks
}

func OpenDM2File(f string) (*DM2File, error) {
	if !u.FileExists(f) {
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

func (demo *DM2File) ParseDM2() error {
	for {
		lump, size, err := nextLump(demo.Handle, int64(demo.Position))
		if size == 0 {
			break
		} else if err != nil {
			return err
		}

		demo.Position += size
		err = demo.ParseLump(lump)
		if err != nil {
			return err
		}
	}
	return nil
}

func nextLump(f *os.File, pos int64) ([]byte, int, error) {
	_, err := f.Seek(pos, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	len := make([]byte, 4)
	_, err = f.Read(len)
	if err != nil {
		return []byte{}, 0, err
	}

	lenbuf := m.MessageBuffer{Buffer: len, Index: 0}
	length := lenbuf.ReadLong()
	if length == -1 {
		return []byte{}, 0, fmt.Errorf("unable to read lump length")
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

func (demo *DM2File) ParseLump(lump []byte) error {
	buf := m.MessageBuffer{Buffer: lump}

	for buf.Index < len(buf.Buffer) {
		cmd := buf.ReadByte()

		switch cmd {
		case m.SVCServerData:
			s := buf.ParseServerData()
			demo.Serverdata = s
			if demo.Callbacks.ServerDataCB != nil {
				demo.Callbacks.ServerDataCB(s)
			}

		case m.SVCConfigString:
			cs := buf.ParseConfigString()
			if !demo.ParsingFrames {
				demo.Configstrings[cs.Index] = cs
			} else {
				demo.CurrentFrame.Strings = append(demo.CurrentFrame.Strings, cs)
			}
			if demo.Callbacks.ConfigStringCB != nil {
				demo.Callbacks.ConfigStringCB(cs)
			}

		case m.SVCSpawnBaseline:
			bl := buf.ParseSpawnBaseline()
			demo.Baselines[bl.Number] = bl
			if demo.Callbacks.BaselineCB != nil {
				demo.Callbacks.BaselineCB(bl)
			}

		case m.SVCStuffText:
			st := buf.ParseStuffText()
			// a "precache" stuff is the delimiter between header data and frames
			if st.String == "precache\n" {
				demo.ParsingFrames = true
			}
			if demo.Callbacks.StuffCB != nil {
				demo.Callbacks.StuffCB(st)
			}

		case m.SVCFrame:
			fr := buf.ParseFrame()
			demo.Frames = append(demo.Frames, m.ServerFrame{})
			if demo.CurrentFrame != nil {
				demo.PrevFrame = demo.CurrentFrame
			}
			demo.CurrentFrame = &demo.Frames[len(demo.Frames)-1]
			demo.CurrentFrame.Frame = fr
			if demo.PrevFrame != nil {
				demo.CurrentFrame.Playerstate = demo.PrevFrame.Playerstate
				demo.CurrentFrame.Entities = demo.PrevFrame.Entities
			}
			if demo.Callbacks.FrameCB != nil {
				demo.Callbacks.FrameCB(fr)
			}

		case m.SVCPlayerInfo:
			ps := buf.ParseDeltaPlayerstate(demo.CurrentFrame.Playerstate)
			demo.CurrentFrame.Playerstate = ps
			if demo.Callbacks.PlayerStateCB != nil {
				demo.Callbacks.PlayerStateCB(ps)
			}

		case m.SVCPacketEntities:
			ents := buf.ParsePacketEntities()
			for _, e := range ents {
				demo.CurrentFrame.Entities[e.Number] = e
			}
			if demo.Callbacks.EntityCB != nil {
				demo.Callbacks.EntityCB(ents)
			}

		case m.SVCPrint:
			p := buf.ParsePrint()
			demo.CurrentFrame.Prints = append(demo.CurrentFrame.Prints, p)
			if demo.Callbacks.PrintCB != nil {
				demo.Callbacks.PrintCB(p)
			}

		case m.SVCSound:
			s := buf.ParseSound()
			demo.CurrentFrame.Sounds = append(demo.CurrentFrame.Sounds, s)
			if demo.Callbacks.SoundCB != nil {
				demo.Callbacks.SoundCB(s)
			}

		case m.SVCTempEntity:
			te := buf.ParseTempEntity()
			demo.CurrentFrame.TempEntities = append(demo.CurrentFrame.TempEntities, te)
			if demo.Callbacks.TempEntCB != nil {
				demo.Callbacks.TempEntCB(te)
			}

		case m.SVCMuzzleFlash:
			mf := buf.ParseMuzzleFlash()
			demo.CurrentFrame.Flash1 = append(demo.CurrentFrame.Flash1, mf)
			if demo.Callbacks.Flash1CB != nil {
				demo.Callbacks.Flash1CB(mf)
			}

		case m.SVCMuzzleFlash2:
			mf := buf.ParseMuzzleFlash()
			demo.CurrentFrame.Flash2 = append(demo.CurrentFrame.Flash2, mf)
			if demo.Callbacks.Flash2CB != nil {
				demo.Callbacks.Flash2CB(mf)
			}

		case m.SVCLayout:
			l := buf.ParseLayout()
			demo.CurrentFrame.Layouts = append(demo.CurrentFrame.Layouts, l)
			if demo.Callbacks.LayoutCB != nil {
				demo.Callbacks.LayoutCB(l)
			}

		case m.SVCInventory:
			buf.ParseInventory()

		case m.SVCCenterPrint:
			c := buf.ParseCenterPrint()
			demo.CurrentFrame.Centerprinters = append(demo.CurrentFrame.Centerprinters, c)
			if demo.Callbacks.CenterPrintCB != nil {
				demo.Callbacks.CenterPrintCB(c)
			}

		default:
			return fmt.Errorf("unknown CMD: %d - %s", cmd, hex.Dump(buf.Buffer[buf.Index-1:]))
		}
	}
	if demo.LumpCallback != nil {
		demo.LumpCallback(lump)
	}
	return nil
}
