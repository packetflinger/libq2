package bsp

import (
	"testing"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

func TestParseEntities(t *testing.T) {
	bsp, err := Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	ents := bsp.GetEntities()
	if got, want := len(ents), 65; got != want {
		t.Fatalf("entity count - have: %d, want: %d", got, want)
	}
	if got, want := ents[0].GetClassName(), "worldspawn"; got != want {
		t.Errorf("first entity classname - have: %q, want: %q", got, want)
	}
	if got, want := ents[0].GetProperties()["message"], "there is no backup .by. deathcubek"; got != want {
		t.Errorf("worldspawn message - have: %q, want: %q", got, want)
	}
	if got, want := ents[1].GetClassName(), "info_player_start"; got != want {
		t.Errorf("second entity classname - have: %q, want: %q", got, want)
	}
	if got, want := ents[1].GetProperties()["origin"], "64 312 408"; got != want {
		t.Errorf("info_player_start origin - have: %q, want: %q", got, want)
	}
}

func TestEntityRoundTrip(t *testing.T) {
	want := []*bpb.BSPEntity{
		{
			ClassName: "worldspawn",
			Properties: map[string]string{
				"classname": "worldspawn",
				"message":   "test map",
			},
		},
		{
			ClassName: "info_player_start",
			Properties: map[string]string{
				"classname": "info_player_start",
				"origin":    "1 2 3",
				"angle":     "90",
			},
		},
	}
	data := marshalEntities(want)
	if len(data) == 0 || data[len(data)-1] != 0 {
		t.Fatal("expected entity lump to end with a trailing NUL byte")
	}
	got, err := parseEntities(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("parsed entity count - have: %d, want: %d", len(got), len(want))
	}
	for i := range want {
		if got[i].GetClassName() != want[i].GetClassName() {
			t.Errorf("entity %d classname - have: %q, want: %q", i, got[i].GetClassName(), want[i].GetClassName())
		}
		for k, v := range want[i].GetProperties() {
			if got[i].GetProperties()[k] != v {
				t.Errorf("entity %d property %q - have: %q, want: %q", i, k, got[i].GetProperties()[k], v)
			}
		}
	}
}

func TestParseEntitiesEmpty(t *testing.T) {
	got, err := parseEntities(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("entity count - have: %d, want: 0", len(got))
	}
}
