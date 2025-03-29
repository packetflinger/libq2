package demo

import (
	"os"

	"github.com/packetflinger/libq2/message"
	pb "github.com/packetflinger/libq2/proto"
)

type MVD2Writer struct {
	demo *pb.MvdDemo
	data message.Buffer
}

// Creates a new writer struct. All the proto-to-binary writing funcs use this
// struct as a receiver.
func NewMVD2Writer(mvd *pb.MvdDemo) *MVD2Writer {
	return &MVD2Writer{
		demo: mvd,
	}
}

// This is the final step when writing a demo. It will write all the binary
// data generated to a file named from the argument.
func (w *MVD2Writer) Finalize(name string) error {
	err := os.WriteFile(name, w.data.Data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// This is the top level function for converting a textproto-based multi-view
// demo to it's proper binary format.
func (w *MVD2Writer) Marshal() error {
	return nil
}

// Generate a binary buffer from a PackedSound proto
func (w *MVD2Writer) MarshalSound(sound *pb.PackedSound) (message.Buffer, error) {
	out := message.Buffer{}
	out.WriteByte(byte(sound.Flags))
	if sound.Index > 255 {
		out.WriteWordP(int32(sound.Index))
	} else {
		out.WriteByteP(sound.Index)
	}
	return out, nil
}

// Generate a binary buffer from a Multicast proto
func (w *MVD2Writer) MarshalMulticast(mc *pb.MvdMulticast) (*message.Buffer, error) {
	out := message.Buffer{}
	out.WriteByteP(uint32(len(mc.Data)))
	if mc.Leaf != 0 {
		out.WriteWordP(int32(mc.Leaf))
	}
	out.WriteData(mc.Data)
	return &out, nil
}
