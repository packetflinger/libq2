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
	demo.currentFrame = &pb.Frame{}
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
			if demo.currentFrame.GetNumber() == 0 {
				textpb.Configstrings = append(textpb.Configstrings, cs)
			} else {
				demo.currentFrame.Configstrings = append(demo.currentFrame.Configstrings, cs)
			}
		case message.SVCSpawnBaseline:
			bitmask := data.ParseEntityBitmask()
			number := data.ParseEntityNumber(bitmask)
			baseline := EntityToProto(&data, bitmask, number)
			textpb.Baselines = append(textpb.Baselines, baseline)
		case message.SVCStuffText:
			stuff := StuffTextToProto(&data)
			if demo.currentFrame.Number > 0 {
				demo.currentFrame.Stufftexts = append(demo.currentFrame.Stufftexts, stuff)
			}
		case message.SVCFrame: // includes playerstate and packetentities
			frame := FrameToProto(&data)
			textpb.Frames = append(textpb.Frames, frame)
			demo.currentFrame = frame
		case message.SVCPrint:
			print := PrintToProto(&data)
			demo.currentFrame.Prints = append(demo.currentFrame.Prints, print)
		case message.SVCMuzzleFlash:
			flash := FlashToProto(&data)
			demo.currentFrame.Flashes1 = append(demo.currentFrame.Flashes1, flash)
		case message.SVCTempEntity:
			te := TempEntToProto(&data)
			demo.currentFrame.TemporaryEntities = append(demo.currentFrame.TemporaryEntities, te)
		case message.SVCLayout:
			layout := LayoutToProto(&data)
			demo.currentFrame.Layouts = append(demo.currentFrame.Layouts, layout)
		case message.SVCSound:
			sound := SoundToProto(&data)
			demo.currentFrame.Sounds = append(demo.currentFrame.Sounds, sound)
		case message.SVCCenterPrint:
			cp := CenterPrintToProto(&data)
			demo.currentFrame.Centerprints = append(demo.currentFrame.Centerprints, cp)
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

	lump.Append(ServerDataToBinary(demo.textProto.Serverinfo))
	for _, cs := range demo.GetTextProto().GetConfigstrings() {
		tmp := ConfigstringToBinary(cs)
		buildDemoBuffer(&out, &lump, tmp, false)
	}
	for _, bl := range demo.GetTextProto().GetBaselines() {
		tmp := message.MessageBuffer{Buffer: []byte{SvcSpawnBaseline}}
		tmp.Append(EntityToBinary(bl))
		buildDemoBuffer(&out, &lump, tmp, false)
	}
	precache := StuffTextToBinary(&pb.StuffText{String_: "precache\n"})
	tmp := message.MessageBuffer{Buffer: []byte{SvcStuffText}}
	tmp.Append(precache)
	buildDemoBuffer(&out, &lump, tmp, false)

	// each frame is a new lump at this point
	for _, fr := range demo.GetTextProto().GetFrames() {
		tmp := FrameToBinary(fr)
		buildDemoBuffer(&out, &lump, tmp, true)
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
