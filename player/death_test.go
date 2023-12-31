package player

import (
	"testing"
)

func TestMOD1(t *testing.T) {
	tests := []struct {
		desc  string
		obit  string
		vic   string
		perp  string
		means int
		solo  bool
	}{
		{
			desc:  "test1",
			obit:  "claire feels scarred's pain",
			vic:   "claire",
			perp:  "scarred",
			means: ModHeldGrenade,
			solo:  true,
		},
		{
			desc:  "test2",
			obit:  "claire was cut in half by shloo's chaingun",
			vic:   "claire",
			perp:  "shloo",
			means: ModChaingun,
			solo:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := CalculateDeath(tc.obit)
			if err != nil {
				t.Error(err)
			}

			if got.Means != tc.means {
				t.Errorf("Means: got %d, want %d\n", got.Means, tc.means)
			}

			if got.Murderer != tc.perp {
				t.Errorf("Murder: got %s, want %s\n", got.Murderer, tc.perp)
			}

			if got.Victim != tc.vic {
				t.Errorf("Victim: got %s, want %s\n", got.Victim, tc.vic)
			}
		})
	}
}
