package demo

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/encoding/prototext"

	pb "github.com/packetflinger/libq2/proto"
)

type DM2Parser struct {
	textProto      *pb.DM2Demo    // uncompressed
	binaryData     []byte         // .dm2 file contents
	binaryPosition int            // where in those contents we are
	currentFrame   int32          // index of frames map
	callbacks      map[int]func() // index is svc_msg type
}

/*
type DM2Writer struct {
	textProto *pb.DM2Demo
}
*/

// Read the entire binary demo file into memory
func NewDM2Demo(filename string) (*DM2Parser, error) {
	if filename == "" {
		return nil, fmt.Errorf("no file specified")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	demo := &DM2Parser{binaryData: data}
	demo.textProto = &pb.DM2Demo{}
	demo.textProto.Baselines = make(map[int32]*pb.PackedEntity)
	demo.textProto.Configstrings = make(map[int32]*pb.ConfigString)
	demo.textProto.Frames = make(map[int32]*pb.Frame)
	return demo, nil
}

// Load the binary demo into protobuf
func (p *DM2Parser) Unmarshal() error {
	for {
		data, length, err := p.NextPacket()
		if err != nil {
			return err
		}
		if length == 0 {
			break
		}
		packet, err := data.ParsePacket(p.textProto.GetFrames())
		if err != nil {
			return err
		}
		err = p.ApplyPacket(packet)
		if err != nil {
			return err
		}
	}
	return nil
}

// ApplyPacket will add all the messages from the packet to the current demo.
func (d *DM2Parser) ApplyPacket(packet *pb.Packet) error {
	if sd := packet.GetServerData(); sd != nil {
		d.textProto.Serverinfo = sd
	}
	cstrings := packet.GetConfigStrings()
	if len(cstrings) > 0 {
		if d.currentFrame > 0 {
			if d.textProto.GetFrames()[d.currentFrame].Configstrings == nil {
				d.textProto.GetFrames()[d.currentFrame].Configstrings = make(map[int32]*pb.ConfigString)
			}
			for _, cs := range cstrings {
				d.textProto.GetFrames()[d.currentFrame].Configstrings[int32(cs.GetIndex())] = cs
			}
		} else {
			for _, cs := range cstrings {
				d.textProto.GetConfigstrings()[int32(cs.GetIndex())] = cs
			}
		}
	}
	baselines := packet.GetBaselines()
	if len(baselines) > 0 {
		for _, bl := range baselines {
			d.textProto.Baselines[int32(bl.GetNumber())] = bl
		}
	}
	frames := packet.GetFrames()
	if len(frames) > 0 {
		// sort them by number ascending
		if len(frames) > 1 {
			slices.SortFunc(frames, func(a, b *pb.Frame) int {
				return cmp.Compare(int(a.GetNumber()), int(b.GetNumber()))
			})
		}
		for _, fr := range frames {
			d.textProto.Frames[int32(fr.GetNumber())] = fr
			d.currentFrame = fr.GetNumber()
		}
	}
	prints := packet.GetPrints()
	if len(prints) > 0 {
		d.textProto.Frames[d.currentFrame].Prints = append(d.textProto.Frames[d.currentFrame].Prints, prints...)
	}
	sounds := packet.GetSounds()
	if len(sounds) > 0 {
		d.textProto.Frames[d.currentFrame].Sounds = append(d.textProto.Frames[d.currentFrame].Sounds, sounds...)
	}
	tempents := packet.GetTempEnts()
	if len(tempents) > 0 {
		d.textProto.Frames[d.currentFrame].TemporaryEntities = append(d.textProto.Frames[d.currentFrame].TemporaryEntities, tempents...)
	}
	mf := packet.GetMuzzleFlashes()
	if len(mf) > 0 {
		d.textProto.Frames[d.currentFrame].Flashes1 = append(d.textProto.Frames[d.currentFrame].Flashes1, mf...)
	}
	layouts := packet.GetLayouts()
	if len(layouts) > 0 {
		d.textProto.Frames[d.currentFrame].Layouts = append(d.textProto.Frames[d.currentFrame].Layouts, layouts...)
	}
	cp := packet.GetCenterprints()
	if len(cp) > 0 {
		d.textProto.Frames[d.currentFrame].Centerprints = append(d.textProto.Frames[d.currentFrame].Centerprints, cp...)
	}
	st := packet.GetStuffs()
	if len(st) > 0 {
		if d.currentFrame > 0 {
			d.textProto.Frames[d.currentFrame].Stufftexts = append(d.textProto.Frames[d.currentFrame].Stufftexts, st...)
		}
	}
	return nil
}

// Demos are organized by "lumps" of data that are essentially packets. Even
// though all the data is already known, each lump represents a server packet's
// worth of game data. Each packet is prefixed with a 32 bit integer of the
// size of the packet and then a bunch of individual messages.
//
// The default packet size for protocol 34 is 1390 bytes. Demos created by
// modern clients using protocols 35/36 will still write demos for protocol
// 34 to maximize compatability. Although it is possible to force these clients
// to record in their native protocol version.
func (demo *DM2Parser) NextPacket() (message.Buffer, int, error) {
	// shouldn't happen, but gracefully handle just in case
	if demo.binaryPosition >= len(demo.binaryData) {
		return message.Buffer{}, 0, errors.New("trying to read past end of packet")
	}
	sizebytes := message.NewBuffer(demo.binaryData[demo.binaryPosition : demo.binaryPosition+4])
	packetLen := int(sizebytes.ReadLong())
	if packetLen == -1 {
		// reached the end of the demo
		return message.Buffer{}, 0, nil
	}
	demo.binaryPosition += 4
	packet := message.Buffer{
		Data: demo.binaryData[demo.binaryPosition : demo.binaryPosition+packetLen],
	}
	demo.binaryPosition += packetLen
	return packet, packetLen, nil
}

// Turn a parsed demo structure back into a binary file
func (demo *DM2Parser) WriteTextProto(filename string) error {
	b, err := prototext.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(demo.textProto)
	if err != nil {
		return fmt.Errorf("error writing proto to file: %s", err.Error())
	}
	err = os.WriteFile(filename, b, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (demo *DM2Parser) GetTextProto() *pb.DM2Demo {
	return demo.textProto
}

// Convert the textproto demo back into a quake 2 playable binary demo. The
// returned byte slice just needs to be written to a file as is.
func (demo *DM2Parser) Marshal() ([]byte, error) {
	out := message.Buffer{}    // the overall demo
	packet := message.Buffer{} // the current packet

	textpb := demo.GetTextProto()

	packet.Append(message.MarshalServerData(textpb.Serverinfo))
	for i := 0; i < MaxConfigStrings; i++ {
		cs, ok := textpb.Configstrings[int32(i)]
		if !ok {
			continue
		}
		tmp := message.MarshalConfigstring(cs)
		buildDemoPacket(&out, &packet, tmp, false)
	}
	for i := 0; i < MaxEdicts; i++ {
		bl, ok := textpb.Baselines[int32(i)]
		if !ok {
			continue
		}
		tmp := message.Buffer{Data: []byte{SvcSpawnBaseline}}
		tmp.Append(message.WriteDeltaEntity(nil, bl))
		buildDemoPacket(&out, &packet, tmp, false)
	}
	tmp := message.Buffer{Data: []byte{SvcStuffText}}
	tmp.Append(message.MarshalStuffText(&pb.StuffText{Data: "precache\n"}))
	buildDemoPacket(&out, &packet, tmp, false)

	frameNum := int32(0)
	total := 0
	for total < len(textpb.GetFrames()) {
		frameNum++
		fr, ok := textpb.Frames[frameNum]
		if !ok {
			continue
		}
		tmp := message.MarshalFrame(fr)
		buildDemoPacket(&out, &packet, tmp, true)
		total++
	}
	out.WriteLong(-1) // end of demo
	return out.Data, nil
}

// Append msg to packet until it can't fit anymore, then append packet to final.
// Each packet is prefixed with its length (4 bytes).
//
// If force is true don't wait until the buffer is full, write it and reset.
//
// Note first two buffer args are pointers since they get updated
func buildDemoPacket(final, packet *message.Buffer, msg message.Buffer, force bool) {
	if ((len(packet.Data) + len(msg.Data)) > message.MaxMessageLength) || force {
		final.WriteLong(int32(len(packet.Data)))
		final.Append(*packet)
		packet.Reset()
	}
	packet.Append(msg)
}

func (p *DM2Parser) RegisterCallback(msgtype int, dofunc func()) {
	p.callbacks[msgtype] = dofunc
}

func (p *DM2Parser) UnregisterCallback(msgtype int) {
	delete(p.callbacks, msgtype)
}
