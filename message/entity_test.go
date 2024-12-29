package message

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseEntityBitmask(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint32
	}{
		{
			name:  "empty input",
			input: "",
			want:  0,
		},
		{
			name:  "1 byte mask",
			input: "100104",
			want:  16,
		},
		{
			name:  "3 byte mask",
			input: "8a3e0100",
			want:  16010,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			b := NewBuffer(h)
			got := b.ParseEntityBitmask()
			if got != tc.want {
				t.Errorf("(%v).ParseEntityBitmask() = %v, want %v\n", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseEntityNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		bitmask uint32
		want    uint16
	}{
		{
			name:    "empty input",
			input:   "",
			bitmask: 0,
			want:    0,
		},
		{
			name:    "single byte number",
			input:   "01040000",
			bitmask: 16,
			want:    1,
		},
		{
			name:    "using EntityNumber16",
			input:   "9903040000",
			bitmask: 16 + 256,
			want:    921,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			b := NewBuffer(h)
			got := b.ParseEntityNumber(tc.bitmask)
			if got != tc.want {
				t.Errorf("(%v).ParseEntityNumber(%v) = %v, want %v\n", tc.input, tc.bitmask, got, tc.want)
			}
		})
	}
}

func TestParseEntity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		from    *pb.PackedEntity
		num     uint16
		bitmask uint32
		want    *pb.PackedEntity
	}{
		{
			name:    "empty everything",
			input:   "",
			from:    nil,
			num:     1,
			bitmask: 16,
			want:    nil,
		},
		{
			name:    "empty from",
			input:   "040000",
			from:    nil,
			num:     1,
			bitmask: 16,
			want: &pb.PackedEntity{
				Number: 1,
				Frame:  4,
			},
		},
		{
			name:  "valid from",
			input: "040000",
			from: &pb.PackedEntity{
				Number:     1,
				ModelIndex: 255,
				Frame:      3,
			},
			num:     1,
			bitmask: 16,
			want: &pb.PackedEntity{
				Number:     1,
				ModelIndex: 255,
				Frame:      4,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			b := NewBuffer(h)
			got := b.ParseEntity(tc.from, tc.num, tc.bitmask)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParseEntity(%v, %v, %v) = %v, want %v\n", tc.input, tc.from, tc.num, tc.bitmask, got, tc.want)
			}
		})
	}
}

func TestParsePacketEntities(t *testing.T) {
	tests := []struct {
		name  string
		input string
		from  map[int32]*pb.PackedEntity
		want  map[int32]*pb.PackedEntity
	}{
		{
			name:  "empty input",
			input: "",
			from:  nil,
			want:  nil,
		},
		{
			name:  "empty from single entity",
			input: "1001040000",
			from:  nil,
			want: map[int32]*pb.PackedEntity{
				1: {Number: 1, Frame: 4},
			},
		},
		{
			name:  "empty from single entity",
			input: "8302018C21FC122D100000",
			from:  nil,
			want: map[int32]*pb.PackedEntity{
				1: {Number: 1, OriginX: 8588, OriginY: 4860, OriginZ: 4141},
			},
		},
		{
			name:  "empty from multiple entities",
			input: "978E900B01FFFF0500019C17090FC118010F9C17090FC1181018001D0023002500260029002B003000310035003600370038003C003D003E003F0040004100420043004400450046004700480049004A004B0050005100520053005400550057005C005D007700790000",
			from:  nil,
			want: map[int32]*pb.PackedEntity{
				1: {
					Number:      1,
					OriginX:     6044,
					OriginY:     3849,
					OriginZ:     6337,
					AngleX:      1,
					AngleY:      15,
					OldOriginX:  6044,
					OldOriginY:  3849,
					OldOriginZ:  6337,
					ModelIndex:  255,
					ModelIndex2: 255,
					Skin:        256,
					Solid:       6160,
					Frame:       5,
				},
				29:  {Number: 29},
				35:  {Number: 35},
				37:  {Number: 37},
				38:  {Number: 38},
				41:  {Number: 41},
				43:  {Number: 43},
				48:  {Number: 48},
				49:  {Number: 49},
				53:  {Number: 53},
				54:  {Number: 54},
				55:  {Number: 55},
				56:  {Number: 56},
				60:  {Number: 60},
				61:  {Number: 61},
				62:  {Number: 62},
				63:  {Number: 63},
				64:  {Number: 64},
				65:  {Number: 65},
				66:  {Number: 66},
				67:  {Number: 67},
				68:  {Number: 68},
				69:  {Number: 69},
				70:  {Number: 70},
				71:  {Number: 71},
				72:  {Number: 72},
				73:  {Number: 73},
				74:  {Number: 74},
				75:  {Number: 75},
				80:  {Number: 80},
				81:  {Number: 81},
				82:  {Number: 82},
				83:  {Number: 83},
				84:  {Number: 84},
				85:  {Number: 85},
				87:  {Number: 87},
				92:  {Number: 92},
				93:  {Number: 93},
				119: {Number: 119},
				121: {Number: 121},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := hex.DecodeString(tc.input)
			if err != nil {
				t.Error(err)
			}
			b := NewBuffer(h)
			got := b.ParsePacketEntities(tc.from)
			if diff := cmp.Diff(got, tc.want, protocmp.Transform()); diff != "" {
				t.Errorf("(%v).ParsePacketEntities(%v) = %v, want %v\n", tc.input, tc.from, got, tc.want)
			}
		})
	}
}

func TestWriteDeltaEntity(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedEntity
		from *pb.PackedEntity
		want string
	}{
		{
			name: "nil entities",
			to:   nil,
			from: nil,
			want: "0000",
		},
		{
			name: "nil from",
			to: &pb.PackedEntity{
				Number:   1,
				RenderFx: 9,
			},
			from: nil,
			want: "80100109",
		},
		{
			name: "valid delta",
			to: &pb.PackedEntity{
				Number:   1,
				RenderFx: 9,
				Event:    10,
			},
			from: &pb.PackedEntity{
				Number:   1,
				RenderFx: 9,
				Event:    0,
			},
			want: "20010a",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := WriteDeltaEntity(tc.from, tc.to)
			got := hex.EncodeToString(buf.Data)
			if got != tc.want {
				t.Errorf("WriteDeltaEntity(%v, %v) = %v, want %v\n", prototext.Format(tc.from), prototext.Format(tc.to), got, tc.want)
			}
		})
	}
}

func TestDeltaEntityBitmask(t *testing.T) {
	tests := []struct {
		name string
		to   *pb.PackedEntity
		from *pb.PackedEntity
		want uint32
	}{
		{
			name: "nil entities",
			to:   nil,
			from: nil,
			want: 0,
		},
		{
			name: "nil from and single byte mask",
			to: &pb.PackedEntity{
				Number:  1,
				OriginX: 5,
				OriginY: 6,
			},
			from: nil,
			want: 3,
		},
		{
			name: "nil from and 2 byte mask",
			to: &pb.PackedEntity{
				Number:  1,
				OriginZ: 5,
				AngleX:  6,
			},
			from: nil,
			want: 1664,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeltaEntityBitmask(tc.to, tc.from)
			if got != tc.want {
				t.Errorf("DeltaEntityBitmask(%v, %v) = %v, want %v\n", prototext.Format(tc.from), prototext.Format(tc.to), got, tc.want)
			}
		})
	}
}
