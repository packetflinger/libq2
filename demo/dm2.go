package demo

import (
	"errors"
	"fmt"
	"os"

	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/encoding/prototext"

	pb "github.com/packetflinger/libq2/proto"
)

type DM2Demo struct {
	textProto      *pb.DM2Demo
	binaryData     []byte // the contents of a .dm2 file
	binaryPosition int    // where in those contents we are
	currentFrame   *pb.Frame
	compressed     bool // every frame contains every edict
	frames         map[int32]*pb.Frame
}

// Read the entire binary demo file into memory
func NewDM2Demo(filename string) (*DM2Demo, error) {
	if filename == "" {
		return nil, fmt.Errorf("no file specified")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	demo := &DM2Demo{binaryData: data}
	demo.textProto = &pb.DM2Demo{}
	demo.textProto.Baselines = make(map[int32]*pb.PackedEntity)
	demo.textProto.Configstrings = make(map[int32]*pb.CString)
	demo.textProto.Frames = make(map[int32]*pb.Frame)

	// demo.currentFrame = &pb.Frame{}
	//demo.frames = make(map[int32]*pb.Frame)
	return demo, nil
}

// Load the binary demo into protobuf
func (demo *DM2Demo) Unmarshal() error {
	for {
		lump, length, err := demo.nextLump()
		if err != nil {
			return err
		}
		if length == 0 {
			break
		}
		err = demo.UnmarshalLump(lump)
		if err != nil {
			return err
		}
	}
	demo.compressed = true
	return nil
}

// Demos are organized by "lumps" of data. Each lump beings with a 32 bit
// integer containing the size of the lump and then a bunch of individual
// messages
func (demo *DM2Demo) nextLump() (message.MessageBuffer, int, error) {
	// shouldn't happen, but gracefully handle just in case
	if demo.binaryPosition >= len(demo.binaryData) {
		return message.MessageBuffer{}, 0, errors.New("trying to read past end of lump")
	}
	sizebytes := message.NewMessageBuffer(demo.binaryData[demo.binaryPosition : demo.binaryPosition+4])
	lumpSize := int(sizebytes.ReadLong())
	if lumpSize == -1 {
		return message.MessageBuffer{}, 0, nil
	}
	demo.binaryPosition += 4
	lump := message.MessageBuffer{
		Buffer: demo.binaryData[demo.binaryPosition : demo.binaryPosition+lumpSize],
	}
	demo.binaryPosition += lumpSize
	return lump, lumpSize, nil
}

// Parse all the messages in a particular chunk of data
func (demo *DM2Demo) UnmarshalLump(data message.MessageBuffer) error {
	textpb := demo.textProto
	for data.Index < len(data.Buffer) {
		cmd := data.ReadByte()
		switch cmd {
		case message.SVCServerData:
			serverdata := ServerDataToProto(&data)
			textpb.Serverinfo = serverdata
		case message.SVCConfigString:
			cs := ConfigstringToProto(&data)
			if textpb.GetCurrentFrame() == 0 {
				textpb.Configstrings[int32(cs.GetIndex())] = cs
			} else {
				textpb.Frames[textpb.GetCurrentFrame()].Configstrings[int32(cs.GetIndex())] = cs
			}
		case message.SVCSpawnBaseline:
			bitmask := data.ParseEntityBitmask()
			number := data.ParseEntityNumber(bitmask)
			baseline := EntityToProto(&data, bitmask, number)
			textpb.Baselines[int32(baseline.GetNumber())] = baseline
		case message.SVCStuffText:
			stuff := StuffTextToProto(&data)
			if textpb.GetCurrentFrame() > 0 {
				textpb.Frames[textpb.GetCurrentFrame()].Stufftexts = append(textpb.Frames[textpb.GetCurrentFrame()].Stufftexts, stuff)
			}
		case message.SVCFrame: // includes playerstate and packetentities
			frame := FrameToProto(&data, demo.frames)
			textpb.Frames[frame.GetNumber()] = frame
			if textpb.Frames[frame.GetNumber()].Configstrings == nil {
				textpb.Frames[frame.GetNumber()].Configstrings = make(map[int32]*pb.CString)
			}
			textpb.CurrentFrame = frame.GetNumber()
		case message.SVCPrint:
			print := PrintToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].Prints = append(textpb.Frames[textpb.GetCurrentFrame()].Prints, print)
		case message.SVCMuzzleFlash:
			flash := FlashToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].Flashes1 = append(textpb.Frames[textpb.GetCurrentFrame()].Flashes1, flash)
		case message.SVCTempEntity:
			te := TempEntToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].TemporaryEntities = append(textpb.Frames[textpb.GetCurrentFrame()].TemporaryEntities, te)
		case message.SVCLayout:
			layout := LayoutToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].Layouts = append(textpb.Frames[textpb.GetCurrentFrame()].Layouts, layout)
		case message.SVCSound:
			sound := SoundToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].Sounds = append(textpb.Frames[textpb.GetCurrentFrame()].Sounds, sound)
		case message.SVCCenterPrint:
			cp := CenterPrintToProto(&data)
			textpb.Frames[textpb.GetCurrentFrame()].Centerprints = append(textpb.Frames[textpb.GetCurrentFrame()].Centerprints, cp)
		}
	}
	return nil
}

// Turn a parsed demo structure back into a binary file
func (demo *DM2Demo) WriteTextProto(filename string) error {
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

func (demo *DM2Demo) GetTextProto() *pb.DM2Demo {
	return demo.textProto
}

// Convert the textproto demo back into a quake 2 playable binary demo. The
// returned byte slice just needs to be written to a file as is.
func (demo *DM2Demo) Marshal() ([]byte, error) {
	out := message.MessageBuffer{}  // the overall demo
	lump := message.MessageBuffer{} // the current lump

	textpb := demo.GetTextProto()

	lump.Append(ServerDataToBinary(textpb.Serverinfo))
	for i := 0; i < MaxConfigStrings; i++ {
		cs, ok := textpb.Configstrings[int32(i)]
		if !ok {
			continue
		}
		tmp := ConfigstringToBinary(cs)
		buildDemoBuffer(&out, &lump, tmp, false)
	}
	for i := 0; i < MaxEdicts; i++ {
		bl, ok := textpb.Baselines[int32(i)]
		if !ok {
			continue
		}
		tmp := message.MessageBuffer{Buffer: []byte{SvcSpawnBaseline}}
		tmp.Append(EntityToBinary(bl))
		buildDemoBuffer(&out, &lump, tmp, false)
	}
	tmp := message.MessageBuffer{Buffer: []byte{SvcStuffText}}
	tmp.Append(StuffTextToBinary(&pb.StuffText{String_: "precache\n"}))
	buildDemoBuffer(&out, &lump, tmp, false)

	i := int32(0)
	total := 0
	for total < len(textpb.GetFrames()) {
		i++
		fr, ok := textpb.Frames[i]
		if !ok {
			continue
		}
		tmp := FrameToBinary(fr)
		buildDemoBuffer(&out, &lump, tmp, true)
		total++
	}
	out.WriteLong(-1) // end of demo
	return out.Buffer, nil
}

// append msg to lump until it can't fit anymore, then append lump to final.
// Each lump is prefixed with its length (4 bytes).
//
// If force is true don't wait until the buffer is full, write it and reset.
func buildDemoBuffer(final *message.MessageBuffer, lump *message.MessageBuffer, msg message.MessageBuffer, force bool) {
	if ((len(lump.Buffer) + len(msg.Buffer)) > message.MaxMessageLength) || force {
		final.WriteLong(int32(len(lump.Buffer)))
		final.Append(*lump)
		lump.Reset()
	}
	lump.Append(msg)
}

// Demos use delta-compression for things like packetentities and playerstates.
// Only property values that are different from the last frame are emitted.
// This saves a huge amount of space and bandwidth because the game doesn't
// have to retransmit the same (unchanged) data over and over again.
//
// Decompressing will ensure every frame contains every entity and playerstate
// property. This will allow for more accurate manipulation of the demo data
// which can then be re-compressed before being written back to a binary file.
func (demo *DM2Demo) Decompress() (*pb.DM2Demo, error) {
	/*
		newdemo := &pb.DM2Demo{}
		info := *demo.textProto.GetServerinfo()
		newdemo.Serverinfo = &info

		configstrings := []*pb.CString{}
		copy(configstrings, demo.textProto.GetConfigstrings())
		newdemo.Configstrings = configstrings

		baselines := []*pb.PackedEntity{}
		copy(baselines, demo.textProto.GetBaselines())
		newdemo.Baselines = baselines
	*/
	//oldframes := make(map[int32]*pb.Frame)
	//mergedFrame := &pb.Frame{}
	//for _, frame := range demo.GetTextProto().GetFrames() {
	/*deltaFrame, ok := oldframes[frame.GetDelta()]
	if !ok {
		ps := *frame.GetPlayerState()
		mergedFrame.PlayerState = &ps
		copy(mergedFrame.Entities, frame.GetEntities())
	} else {

	}
	oldframes[frame.GetNumber()] = frame
	newdemo.Frames = append(newdemo.Frames, mergedFrame)*/
	//}

	return &pb.DM2Demo{}, nil
}

/*
func mergeFrame(base, current *pb.Frame) *pb.Frame {
	fr := proto.Clone(base).(*pb.Frame)
	fr.Number = current.GetNumber()
	fr.Delta = current.GetDelta()

	return fr
}

func mergePlayerstate(base, current *pb.PackedPlayer) *pb.PackedPlayer {
	ms := proto.Clone(base.Movestate).(*pb.PlayerMove)
	if ms.GetType() != current.GetMovestate().GetType() {
		base.Movestate.Type = current.Movestate.Type
	}
	if ms.GetOriginX() != current.GetMovestate().GetOriginX() {
		base.Movestate.OriginX = current.Movestate.OriginX
	}
	if ms.GetOriginY() != current.GetMovestate().GetOriginY() {
		base.Movestate.OriginY = current.Movestate.OriginY
	}
	if ms.GetOriginZ() != current.GetMovestate().GetOriginZ() {
		base.Movestate.OriginZ = current.Movestate.OriginZ
	}
	if ms.GetVelocityX() != current.GetMovestate().GetVelocityX() {
		base.Movestate.VelocityX = current.Movestate.VelocityX
	}
	if ms.GetVelocityY() != current.GetMovestate().GetVelocityY() {
		base.Movestate.VelocityY = current.Movestate.VelocityY
	}
	if ms.GetVelocityZ() != current.GetMovestate().GetVelocityZ() {
		base.Movestate.VelocityZ = current.Movestate.VelocityZ
	}
	if ms.GetFlags() != current.GetMovestate().GetFlags() {
		base.Movestate.Flags = current.Movestate.Flags
	}
	if ms.GetTime() != current.GetMovestate().GetTime() {
		base.Movestate.Time = current.Movestate.Time
	}
	if ms.GetGravity() != current.GetMovestate().GetGravity() {
		base.Movestate.Gravity = current.Movestate.Gravity
	}
	if ms.GetDeltaAngleX() != current.GetMovestate().GetDeltaAngleX() {
		base.Movestate.DeltaAngleX = current.Movestate.DeltaAngleX
	}
	if ms.GetDeltaAngleY() != current.GetMovestate().GetDeltaAngleY() {
		base.Movestate.DeltaAngleY = current.Movestate.DeltaAngleY
	}
	if ms.GetDeltaAngleZ() != current.GetMovestate().GetDeltaAngleZ() {
		base.Movestate.DeltaAngleZ = current.Movestate.DeltaAngleZ
	}
	if base.GetViewAnglesX() != current.GetViewAnglesX() {
		base.ViewAnglesX = current.ViewAnglesX
	}
	if base.GetViewAnglesY() != current.GetViewAnglesY() {
		base.ViewAnglesY = current.ViewAnglesY
	}
	if base.GetViewAnglesZ() != current.GetViewAnglesZ() {
		base.ViewAnglesZ = current.ViewAnglesZ
	}
	if base.GetViewOffsetX() != current.GetViewOffsetX() {
		base.ViewOffsetX = current.ViewOffsetX
	}
	if base.GetViewAnglesY() != current.GetViewAnglesY() {
		base.ViewAnglesY = current.ViewAnglesY
	}
	if base.GetViewAnglesZ() != current.GetViewAnglesZ() {
		base.ViewAnglesZ = current.ViewAnglesZ
	}
	return base
}
*/
