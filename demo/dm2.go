package demo

import (
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
	return demo, nil
}

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
	currentFrame := &pb.Frame{}
	textpb := demo.textProto
	for data.Index < len(data.Buffer) {
		cmd := data.ReadByte()
		switch cmd {
		case message.SVCServerData:
			serverdata := ServerDataToProto(&data)
			textpb.Serverinfo = serverdata
		case message.SVCConfigString:
			cs := ConfigstringToProto(&data)
			if currentFrame.GetNumber() == 0 {
				textpb.Configstrings = append(textpb.Configstrings, cs)
			} else {
				currentFrame.Configstrings = append(currentFrame.Configstrings, cs)
			}
		case message.SVCSpawnBaseline:
			bitmask := data.ParseEntityBitmask()
			number := data.ParseEntityNumber(bitmask)
			baseline := EntityToProto(&data, bitmask, number)
			textpb.Baselines = append(textpb.Baselines, baseline)
		case message.SVCStuffText:
			stuff := StuffTextToProto(&data)
			if currentFrame.Number > 0 {
				currentFrame.Stufftexts = append(currentFrame.Stufftexts, stuff)
			}
		case message.SVCFrame: // includes playerstate and packetentities
			frame := FrameToProto(&data)
			textpb.Frames = append(textpb.Frames, frame)
			currentFrame = frame
		case message.SVCPrint:
			print := PrintToProto(&data)
			currentFrame.Prints = append(currentFrame.Prints, print)
		case message.SVCMuzzleFlash:
			flash := FlashToProto(&data)
			currentFrame.Flashes1 = append(currentFrame.Flashes1, flash)
		case message.SVCTempEntity:
			te := TempEntToProto(&data)
			currentFrame.TemporaryEntities = append(currentFrame.TemporaryEntities, te)
		case message.SVCLayout:
			layout := LayoutToProto(&data)
			currentFrame.Layouts = append(currentFrame.Layouts, layout)
		case message.SVCSound:
			sound := SoundToProto(&data)
			currentFrame.Sounds = append(currentFrame.Sounds, sound)
		case message.SVCCenterPrint:
			cp := CenterPrintToProto(&data)
			currentFrame.Centerprints = append(currentFrame.Centerprints, cp)
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
