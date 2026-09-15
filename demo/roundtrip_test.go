package demo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/packetflinger/libq2/message"
	pb "github.com/packetflinger/libq2/proto"
)

// TestMvdRoundTrip verifies that a real captured demo, once parsed, can be
// re-marshaled and re-parsed to produce the same logical demo data. This is
// the strongest correctness check available for the writer without a
// reference recording from q2pro itself to diff against byte-for-byte.
func TestMvdRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		demofile string
	}{
		{name: "test1", demofile: "../testdata/test.mvd2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewMVD2Parser(tc.demofile)
			if err != nil {
				t.Fatalf("error creating parser: %v", err)
			}
			demos, err := parser.Unmarshal()
			if err != nil {
				t.Fatalf("error unmarshalling: %v", err)
			}
			if len(demos) == 0 {
				t.Fatal("no demos parsed")
			}

			// Re-marshal every embedded demo (map) back into one continuous
			// stream, exactly like a real multi-map recording.
			out := message.NewBuffer(nil)
			out.WriteLong(MVDMagic)
			for _, d := range demos {
				writer := NewMVD2Writer(&pb.MvdDemo{Packets: d.Packets})
				for _, pkt := range d.Packets {
					body := writer.MarshalPacket(pkt)
					out.WriteShort(body.Size())
					out.Append(body)
				}
			}
			out.WriteShort(0)

			tmp := filepath.Join(t.TempDir(), "roundtrip.mvd2")
			if err := os.WriteFile(tmp, out.Data, 0644); err != nil {
				t.Fatalf("error writing temp file: %v", err)
			}

			reparsed, err := NewMVD2Parser(tmp)
			if err != nil {
				t.Fatalf("error re-parsing round-tripped file: %v", err)
			}
			redemos, err := reparsed.Unmarshal()
			if err != nil {
				t.Fatalf("error re-unmarshalling round-tripped file: %v", err)
			}

			if len(redemos) != len(demos) {
				t.Fatalf("got %d demos after round trip, want %d", len(redemos), len(demos))
			}
			// old_origin on a brand new entity is only sent when the
			// original encoder decided the entity needed it (e.g. a mover
			// with RF_FRAMELERP/RF_BEAM); the parsed proto doesn't preserve
			// that decision, so the writer can only approximate it
			// heuristically. Everything else should match exactly.
			//
			// Comparisons run packet-by-packet (rather than diffing the
			// whole demo at once) so a real captured file with tens of
			// thousands of packets stays fast and gives a useful report
			// instead of one enormous diff.
			ignore := protocmp.IgnoreFields(&pb.PackedEntity{},
				"old_origin_x", "old_origin_y", "old_origin_z")
			for i := range demos {
				if len(demos[i].GetPackets()) != len(redemos[i].GetPackets()) {
					t.Fatalf("demo %d: got %d packets after round trip, want %d",
						i, len(redemos[i].GetPackets()), len(demos[i].GetPackets()))
				}
				mismatches := 0
				for j := range demos[i].GetPackets() {
					diff := cmp.Diff(demos[i].Packets[j], redemos[i].Packets[j], protocmp.Transform(), ignore)
					if diff == "" {
						continue
					}
					mismatches++
					if mismatches <= 3 {
						t.Errorf("demo %d packet %d round trip mismatch:\n%s", i, j, diff)
					}
				}
				if mismatches > 3 {
					t.Errorf("demo %d: %d packets mismatched in total", i, mismatches)
				}
			}
		})
	}
}
