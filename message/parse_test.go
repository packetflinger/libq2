package message

/*
func TestValidateSequence(t *testing.T) {
	tests := []struct {
		desc         string
		seq          uint32
		wantreliable bool
		wantsequence uint32
	}{
		{
			desc:         "Test_1 pass",
			seq:          5,
			wantreliable: false,
			wantsequence: 5,
		},
		{
			desc:         "Test_2 pass",
			seq:          2147483652,
			wantreliable: true,
			wantsequence: 4,
		},
	}

	for _, test := range tests {
		r, s := ValidateSequence(test.seq)
		if r != test.wantreliable {
			t.Errorf("%s ValidateSequnce() reliability mismatch - have %t, want %t\n", test.desc, r, test.wantreliable)
		}

		if s != test.wantsequence {
			t.Errorf("%s ValidateSequnce() sequence mismatch - have %d, want %d\n", test.desc, s, test.wantsequence)
		}
	}
}
*/
