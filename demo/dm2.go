// Quake 2 demos are essentially just a copy of the network stream from the
// server to a client; a one-sided conversation of network messages. Client to
// server messages are irrelevant as the server will process those messages and
// transmit the results to all clients anyway.
//
// All .dm2 demo files are recorded using protocol 34 regardless of what
// protocol the client recording is connected at.
//
// Demos are grouped into "lumps" of data that could be thought of as packets.
// Each packet is prefixed with a 4-byte integer size indicating the length
// of that particular packet. Then each packet is just a collection of server
// messages. The header will be a few packets containing a `serverdata` message
// followed by all the initial configstrings and then all the baselines. At
// that point there'll be stufftext message containing the string "precache\n"
// and then the client is considered spawned into the game and serverframes
// will start flowing. Serverframes are always followed by a playerstate msg
// and then all the packetentities that changed for that particular frame;
// always in that order. The other messages (sounds, temp entities, etc) can
// come before and after, but those three will always be together and in that
// order once per frame.
package demo

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/packetflinger/libq2/message"

	pb "github.com/packetflinger/libq2/proto"
)

type DM2Parser struct {
	textProto      *pb.DM2Demo       // uncompressed
	binaryData     []byte            // .dm2 file contents
	binaryPosition int               // where in those contents we are
	currentFrame   int32             // index of frames map
	callbacks      map[int]func(any) // index is svc_msg type
	frameCount     int               // how many frames in total
	fps            int               // only supports 10
}

// Create a new DM2 parser to hold a parsed demo.
func NewDM2Parser() *DM2Parser {
	return &DM2Parser{
		callbacks:  make(map[int]func(any)),
		frameCount: 0,
		fps:        10,
		textProto: &pb.DM2Demo{
			Baselines:     make(map[int32]*pb.PackedEntity),
			Configstrings: make(map[int32]*pb.ConfigString),
			Frames:        make(map[int32]*pb.Frame),
		},
	}
}

// Load the binary demo into protobuf
func (p *DM2Parser) Unmarshal(content []byte) error {
	if len(content) == 0 {
		return fmt.Errorf("empty demo data")
	}
	p.binaryData = content
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
		if cbFunc, found := d.callbacks[message.SVCConfigString]; found {
			for _, cs := range cstrings {
				cbFunc(cs)
			}
		}
	}
	baselines := packet.GetBaselines()
	if len(baselines) > 0 {
		for _, bl := range baselines {
			d.textProto.Baselines[int32(bl.GetNumber())] = bl
		}
		if cbFunc, found := d.callbacks[message.SVCSpawnBaseline]; found {
			for _, bl := range baselines {
				cbFunc(bl)
			}
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
			d.frameCount++
		}
		if cbFunc, found := d.callbacks[message.SVCFrame]; found {
			for _, fr := range frames {
				cbFunc(fr)
			}
		}
	}
	prints := packet.GetPrints()
	if len(prints) > 0 {
		d.textProto.Frames[d.currentFrame].Prints = append(d.textProto.Frames[d.currentFrame].Prints, prints...)
		if cbFunc, found := d.callbacks[message.SVCPrint]; found {
			for _, pr := range prints {
				cbFunc(pr)
			}
		}
	}
	sounds := packet.GetSounds()
	if len(sounds) > 0 {
		d.textProto.Frames[d.currentFrame].Sounds = append(d.textProto.Frames[d.currentFrame].Sounds, sounds...)
		if cbFunc, found := d.callbacks[message.SVCSound]; found {
			for _, snd := range sounds {
				cbFunc(snd)
			}
		}
	}
	tempents := packet.GetTempEnts()
	if len(tempents) > 0 {
		d.textProto.Frames[d.currentFrame].TemporaryEntities = append(d.textProto.Frames[d.currentFrame].TemporaryEntities, tempents...)
		if cbFunc, found := d.callbacks[message.SVCTempEntity]; found {
			for _, te := range tempents {
				cbFunc(te)
			}
		}
	}
	mf := packet.GetMuzzleFlashes()
	if len(mf) > 0 {
		d.textProto.Frames[d.currentFrame].Flashes1 = append(d.textProto.Frames[d.currentFrame].Flashes1, mf...)
		if cbFunc, found := d.callbacks[message.SVCMuzzleFlash]; found {
			for _, f := range mf {
				cbFunc(f)
			}
		}
	}
	layouts := packet.GetLayouts()
	if len(layouts) > 0 {
		d.textProto.Frames[d.currentFrame].Layouts = append(d.textProto.Frames[d.currentFrame].Layouts, layouts...)
		if cbFunc, found := d.callbacks[message.SVCLayout]; found {
			for _, lo := range layouts {
				cbFunc(lo)
			}
		}
	}
	cp := packet.GetCenterprints()
	if len(cp) > 0 {
		d.textProto.Frames[d.currentFrame].Centerprints = append(d.textProto.Frames[d.currentFrame].Centerprints, cp...)
		if cbFunc, found := d.callbacks[message.SVCCenterPrint]; found {
			for _, c := range cp {
				cbFunc(c)
			}
		}
	}
	st := packet.GetStuffs()
	if len(st) > 0 {
		if d.currentFrame > 0 {
			d.textProto.Frames[d.currentFrame].Stufftexts = append(d.textProto.Frames[d.currentFrame].Stufftexts, st...)
		}
		if cbFunc, found := d.callbacks[message.SVCStuffText]; found {
			for _, s := range st {
				cbFunc(s)
			}
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
	packet := message.NewBuffer(demo.binaryData[demo.binaryPosition : demo.binaryPosition+packetLen])
	demo.binaryPosition += packetLen
	return packet, packetLen, nil
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
	for i := range MaxConfigStrings {
		cs, ok := textpb.Configstrings[int32(i)]
		if !ok {
			continue
		}
		tmp := message.MarshalConfigstring(cs)
		buildDemoPacket(&out, &packet, tmp, false)
	}
	for i := range MaxEdicts {
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
		final.WriteLong(len(packet.Data))
		final.Append(*packet)
		packet.Reset()
	}
	packet.Append(msg)
}

// RegisterCallback allows for a custom function to be called at specific
// points while a demo is being parsed.
//
// `event` is an index corresponding to what you want your callback to be
// associated with. These match up with the SVC* server messages.
//
// `dofunc` is the function definition to be called at that event. The argument
// is set to `any` to accept any type, but this arg should be the parsed server
// message proto. Inside the dofunc, the arg will need to be casted to the
// appropriate type. Something like:
//
// printMsg := argname.(*pb.Print)
func (p *DM2Parser) RegisterCallback(event int, dofunc func(any)) {
	p.callbacks[event] = dofunc
}

// Dynamically remove a particular callback
func (p *DM2Parser) UnregisterCallback(msgtype int) {
	delete(p.callbacks, msgtype)
}
