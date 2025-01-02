package message

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	pb "github.com/packetflinger/libq2/proto"
)

func hexToBuffer(in string) (Buffer, error) {
	input := strings.Replace(in, " ", "", -1) // remove spaces
	bytes, err := hex.DecodeString(input)
	if err != nil {
		return Buffer{}, err
	}
	b := NewBuffer(bytes)
	return b, nil
}

func bufferToHex(in Buffer) string {
	return hex.EncodeToString(in.Data)
}

func TestParseServerData(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.ServerInfo
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid serverdata from demo",
			data: "220000007F4CED3001000000546865204564676500",
			want: &pb.ServerInfo{
				Protocol:     34,
				ServerCount:  820857983,
				Demo:         true,
				GameDir:      "",
				ClientNumber: 0,
				MapName:      "The Edge",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseServerData()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseServerData() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseSpawnBaseline(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.PackedEntity
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid baseline",
			data: "838280047DA0FF0012E0172C0E838280047E6017A002E01F2C0E838280047FE02EA01E200D2B",
			want: &pb.PackedEntity{
				Number:  125,
				OriginX: 65440,
				OriginY: 4608,
				OriginZ: 6112,
				Sound:   44,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseSpawnBaseline()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseServerSpawnBaseline() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseStuffText(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.StuffText
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid stuff",
			data: "70726563616368650A00",
			want: &pb.StuffText{
				Data: "precache\n",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseStuffText()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseStuffText() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseFrame(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		oldframes map[int32]*pb.PackedEntity
		want      *pb.Frame
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid frame",
			data: "1D0000001C00000000010211000000000000121001090000",
			want: &pb.Frame{
				Number:    29,
				Delta:     28,
				AreaBytes: 1,
				AreaBits:  []uint32{2},
				PlayerState: &pb.PackedPlayer{
					Movestate: &pb.PlayerMove{},
				},
				Entities: map[int32]*pb.PackedEntity{
					1: {
						Number: 1,
						Frame:  9,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseFrame(nil)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseFrame() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParsePrint(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.Print
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid print",
			data: "03636C616972653A203A29200A00",
			want: &pb.Print{
				Level: 3,
				Data:  "claire: :) \n",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParsePrint()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParsePrint() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

/*func TestParseSound(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.PackedSound
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid sound",
			data: "03636C616972653A203A29200A00",
			want: &pb.PackedSound{
				Flags: 2,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParsePrint()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParsePrint() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}*/

func TestParseConfigString(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.ConfigString
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid configstring",
			data: "0000546865204564676500",
			want: &pb.ConfigString{
				Index: 0,
				Data:  "The Edge",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := hex.DecodeString(tc.data)
			in := NewBuffer(b)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseConfigString()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseConfigString() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseChallenge(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    *pb.Challenge
		wantErr bool
	}{
		{
			name: "valid all three protocols",
			data: "ffffffff6368616c6c656e67652039313039303836343420703d33342c33352c3336",
			want: &pb.Challenge{
				Number:    910908644,
				Protocols: []int32{34, 35, 36},
			},
			wantErr: false,
		},
		{
			name: "valid only two protocols",
			data: "ffffffff6368616c6c656e67652032363832393035313320703d33342c3335",
			want: &pb.Challenge{
				Number:    268290513,
				Protocols: []int32{34, 35},
			},
			wantErr: false,
		},
		{
			name: "valid no protocols",
			data: "ffffffff6368616c6c656e676520323638323930353133",
			want: &pb.Challenge{
				Number:    268290513,
				Protocols: []int32{},
			},
			wantErr: false,
		},
		{
			name:    "invalid sequence",
			data:    "6368616c6c656e67652032363832393035313320703d33342c3335",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := hex.DecodeString(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			msg := NewBuffer(bytes)
			got, err := msg.ParseChallenge()
			if err != nil && !tc.wantErr {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("ParseChallenge() error (-got/+want):\n%v\n", diff)
			}
		})
	}
}

func TestParseTempEntity(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.TemporaryEntity
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid tempent",
			data: "07682C BD192A0E",
			want: &pb.TemporaryEntity{
				Type:       7,
				Position1X: 11368,
				Position1Y: 6589,
				Position1Z: 3626,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hexToBuffer(tc.data)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseTempEntity()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseTempEntity() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseMuzzleFlash(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.MuzzleFlash
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid flash",
			data: "010007",
			want: &pb.MuzzleFlash{
				Entity: 1,
				Weapon: 7,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hexToBuffer(tc.data)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseMuzzleFlash()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseMuzzleFlash() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseLayout(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.Layout
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid layout",
			data: "546865204564676500",
			want: &pb.Layout{
				Data: "The Edge",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hexToBuffer(tc.data)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseLayout()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseLayout() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParseCenterPrint(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *pb.Layout
	}{
		{
			name: "empty data",
			data: "",
			want: nil,
		},
		{
			name: "valid centerprint",
			data: "546865204564676500",
			want: &pb.Layout{
				Data: "The Edge",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hexToBuffer(tc.data)
			if err != nil {
				t.Error(err)
			}
			got := in.ParseLayout()
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseLayout() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestParsePacket(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		oldframes map[int32]*pb.Frame
		want      *pb.Packet
	}{
		{
			name:      "empty data",
			data:      "",
			oldframes: nil,
			want:      nil,
		},
		{
			name:      "valid packet",
			data:      "14140000 00130000 00000102 1100200E F400F600 00000000 00001210 01004080 00000307 682CBD19 2A0E",
			oldframes: nil,
			want: &pb.Packet{
				Frames: []*pb.Frame{
					{
						Number:    20,
						Delta:     19,
						AreaBytes: 1,
						AreaBits:  []uint32{2},
						PlayerState: &pb.PackedPlayer{
							Movestate:  &pb.PlayerMove{},
							GunOffsetX: -12,
							GunOffsetZ: -10,
							GunFrame:   14,
						},
						Entities: map[int32]*pb.PackedEntity{
							1:   {Number: 1},
							128: {Number: 128, Remove: true},
						},
					},
				},
				TempEnts: []*pb.TemporaryEntity{
					{
						Type:       7,
						Position1X: 11368,
						Position1Y: 6589,
						Position1Z: 3626,
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in, err := hexToBuffer(tc.data)
			if err != nil {
				t.Error(err)
			}
			got, err := in.ParsePacket(tc.oldframes)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseLayout() resulted in diff:\n%v", tc.data, diff)
			}
		})
	}
}

func TestMarshalServerData(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.ServerInfo
		want string
	}{
		{
			name: "empty data",
			in:   nil,
			want: "0c00000000000000000000000000",
		},
		{
			name: "valid serverdata",
			in: &pb.ServerInfo{
				Protocol:     34,
				ServerCount:  820857983,
				Demo:         true,
				GameDir:      "",
				ClientNumber: 0,
				MapName:      "The Edge",
			},
			want: "0c220000007F4CED3001000000546865204564676500",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalServerData(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalServerData(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalCenterprint(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.CenterPrint
		want string
	}{
		{
			name: "test 1",
			in: &pb.CenterPrint{
				Data: "The Edge",
			},
			want: "546865204564676500",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalCenterPrint(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalCenterprint(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalConfigstring(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.ConfigString
		want string
	}{
		{
			name: "test 1",
			in: &pb.ConfigString{
				Index: 5,
				Data:  "The Edge",
			},
			want: "0d0500546865204564676500",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalConfigstring(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalConfigstring(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalFlash(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.MuzzleFlash
		want string
	}{
		{
			name: "test 1",
			in: &pb.MuzzleFlash{
				Entity: 1,
				Weapon: 7,
			},
			want: "010007",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalFlash(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalFlash(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalFrame(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.Frame
		want string
	}{
		{
			name: "test 1",
			in: &pb.Frame{
				Number:    29,
				Delta:     28,
				AreaBytes: 1,
				AreaBits:  []uint32{2},
				PlayerState: &pb.PackedPlayer{
					Movestate: &pb.PlayerMove{},
				},
				Entities: map[int32]*pb.PackedEntity{
					1: {
						Number: 1,
						Frame:  9,
					},
				},
			},
			want: "141D0000001C00000000010211000000000000121001090000",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalFrame(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalFrame(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalPrint(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.Print
		want string
	}{
		{
			name: "test 1",
			in: &pb.Print{
				Level: 3,
				Data:  "The Edge",
			},
			want: "03546865204564676500",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalPrint(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalPrint(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalStuffText(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.StuffText
		want string
	}{
		{
			name: "test 1",
			in: &pb.StuffText{
				Data: "The Edge",
			},
			want: "546865204564676500",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalStuffText(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalStuffText(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarshalTempEntity(t *testing.T) {
	tests := []struct {
		name string
		in   *pb.TemporaryEntity
		want string
	}{
		{
			name: "test 1",
			in: &pb.TemporaryEntity{
				Type:       7,
				Position1X: 11368,
				Position1Y: 6589,
				Position1Z: 3626,
			},
			want: "07682CBD192A0E",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bufferToHex(MarshalTempEntity(tc.in))
			if !strings.EqualFold(got, tc.want) {
				t.Errorf("MarshalTempEntity(%v) = %v, want %v\n", tc.in, got, tc.want)
			}
		})
	}
}
