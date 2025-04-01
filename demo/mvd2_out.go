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

func (w *MVD2Writer) MarshalServerData(data *pb.MvdServerData) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteLongP(data.GetVersionMajor())
	out.WriteShortP(data.GetVersionMinor())
	if data.GetVersionMinor() >= ProtocolPlusPlus {
		out.WriteShortP(w.demo.GetFlags())
	}
	out.WriteLongP(data.GetSpawnCount())
	out.WriteString(data.GetGameDir())
	out.WriteShortP(data.GetClientNumber())
	return out
}

func (w *MVD2Writer) MarshalConfigstrings(data map[int32]*pb.ConfigString) message.Buffer {
	out := message.NewBuffer(nil)
	for _, cs := range data {
		out.Append(w.MarshalConfigstring(cs))
	}
	out.WriteShortP(w.demo.GetRemap().GetEnd())
	return out
}

func (w *MVD2Writer) MarshalConfigstring(data *pb.ConfigString) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteShortP(int32(data.GetIndex()))
	out.WriteString(data.GetData())
	return out
}

// Generate a binary buffer from a PackedSound proto
func (w *MVD2Writer) MarshalSound(sound *pb.PackedSound) (message.Buffer, error) {
	out := message.Buffer{}
	out.WriteByte(int(sound.Flags))
	if sound.Index > 255 {
		out.WriteWordP(sound.Index)
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
		out.WriteWordP(uint32(mc.Leaf))
	}
	out.WriteData(mc.Data)
	return &out, nil
}
