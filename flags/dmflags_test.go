package flags

import "testing"

func TestModify(t *testing.T) {
	tests := []struct {
		name  string
		flags int
		instr string
		want  int
	}{
		{
			name:  "no instruction",
			flags: QuadDrop | InstantItems | NoFalling,
			instr: "",
			want:  QuadDrop | InstantItems | NoFalling,
		},
		{
			name:  "bad instruction",
			flags: QuadDrop | InstantItems | NoFalling,
			instr: "jello!",
			want:  QuadDrop | InstantItems | NoFalling,
		},
		{
			name:  "mixed good and bad instructions",
			flags: QuadDrop | InstantItems | NoFalling,
			instr: "+sl +boner -qd",
			want:  SameLevel | InstantItems | NoFalling,
		},
		{
			name:  "add something",
			flags: QuadDrop,
			instr: "+ia",
			want:  InfiniteAmmo | QuadDrop,
		},
		{
			name:  "remove falling damage",
			flags: QuadDrop | NoArmor | InstantItems,
			instr: "-fd",
			want:  QuadDrop | NoArmor | InstantItems | NoFalling,
		},
		{
			name:  "remove something",
			flags: QuadDrop | NoHealth | InstantItems,
			instr: "-nh",
			want:  QuadDrop | InstantItems,
		},
		{
			name:  "add, rmove, add existing",
			flags: QuadDrop | WeaponsStay | InstantItems,
			instr: "-ws +ia +qd",
			want:  QuadDrop | InstantItems | InfiniteAmmo,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Modify(tc.flags, tc.instr)
			if got != tc.want {
				t.Errorf("Modify(%d, %q) = %d, want %d\n", tc.flags, tc.instr, got, tc.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name  string
		flags int
		want  string
	}{
		{
			name:  "zero",
			flags: 0,
			want:  "",
		},
		{
			name:  "negative",
			flags: -30000,
			want:  "fixed FOV, instant items, no armor, teams by model, teams by skin",
		},
		{
			name:  "single",
			flags: QuadDrop,
			want:  "quad drop",
		},
		{
			name:  "common server",
			flags: FreeForAllServer,
			want:  "force respawn, instant items, weapons stay",
		},
		{
			name:  "too big",
			flags: 65536,
			want:  "",
		},
		{
			name:  "common",
			flags: 1040,
			want:  "force respawn, instant items",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ToString(tc.flags)
			if got != tc.want {
				t.Errorf("ToString(%v) = %q, want %q\n", tc.flags, got, tc.want)
			}
		})
	}
}
