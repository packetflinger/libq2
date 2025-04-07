package demo

import (
	"os"
	"slices"

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

func (w *MVD2Writer) MarshalServerData() message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteLongP(37)
	out.WriteShortP(w.demo.GetVersion())
	if w.demo.GetVersion() >= ProtocolPlusPlus {
		out.WriteShortP(w.demo.GetFlags())
	}
	out.WriteLongP(w.demo.GetIdentity())
	out.WriteString(w.demo.GetGameDir())
	out.WriteShortP(w.demo.GetDummy())
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

func (w *MVD2Writer) MarshalFrame(frame *pb.MvdFrame) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteByte(len(frame.GetPortalData()))
	out.WriteData(frame.GetPortalData())
	out.Append(w.MarshalPlayers(frame.GetPlayers()))
	return out
}

func (w *MVD2Writer) MarshalPlayers(players map[int32]*pb.PackedPlayer) message.Buffer {
	out := message.NewBuffer(nil)
	for num, pl := range players {
		out.Append(w.MarshalPlayer(num, pl))
	}
	out.WriteByte(ClientNumNone)
	return out
}

func (w *MVD2Writer) MarshalPlayer(num int32, player *pb.PackedPlayer) message.Buffer {
	out := message.NewBuffer(nil)

	from := w.demo.GetPlayers()[num]
	bits := message.DeltaPlayerBitmask(from.GetPlayerState(), player)

	out.WriteByte(int(num))
	out.WriteWord(bits)
	out.Append(message.WriteDeltaPlayerstate(from.GetPlayerState(), player))
	return out
}

func (w *MVD2Writer) MarshalEntities(ents map[int32]*pb.PackedEntity) message.Buffer {
	out := message.NewBuffer(nil)

	// ents need to be in numeric order and maps are not guaranteed to give
	// their values in the order they were added. So export the keys and
	// sort them.
	var keys []int32
	for k := range ents {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		out.Append(message.WriteDeltaEntity(ents[k], w.demo.GetEntities()[k]))
	}
	out.WriteShort(0) // combined bitmask and number
	return out
}
