package message

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

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
