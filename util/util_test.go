package util

import (
	"testing"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		desc     string
		value    int
		upper    int
		lower    int
		want     int
		wantPass bool
	}{
		{
			desc:     "test_1 - pass, inside",
			value:    50,
			upper:    55,
			lower:    45,
			want:     50,
			wantPass: true,
		},
		{
			desc:     "test_2 - pass, lower",
			value:    4,
			upper:    55,
			lower:    45,
			want:     45,
			wantPass: true,
		},
		{
			desc:     "test_3 - pass, upper",
			value:    500,
			upper:    55,
			lower:    45,
			want:     55,
			wantPass: true,
		},
		{
			desc:     "test_4 - fail, inside",
			value:    50,
			upper:    55,
			lower:    45,
			want:     100,
			wantPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := Clamp(tc.value, tc.lower, tc.upper)
			if (got == tc.want) != tc.wantPass {
				t.Errorf("Clamp fail, got %d, want %d\n", got, tc.want)
			}
		})
	}
}

func TestConvertHighChars(t *testing.T) {
	in := "áîôéãèåáô"
	want := "anticheat"
	got := ConvertHighChars(in)
	if got != want {
		t.Error("ConvertHighChars: got", got, "want", want)
	}
}

func TestConvertLowChars(t *testing.T) {
	in := "anticheat"
	want := "áîôéãèåáô"
	got := ConvertLowChars(in)
	if got != want {
		t.Error("ConvertLowChars: got", got, "want", want)
	}
}
