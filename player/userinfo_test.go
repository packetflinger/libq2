package player

import "testing"

func TestUserinfoAdd(t *testing.T) {
	tests := []struct {
		name    string
		ui      Userinfo
		key     string
		value   string
		wantErr bool
	}{
		{
			name:    "both empty",
			ui:      NewUserinfo(),
			key:     "",
			value:   "",
			wantErr: true,
		},
		{
			name:    "key empty",
			ui:      NewUserinfo(),
			key:     "",
			value:   "val",
			wantErr: true,
		},
		{
			name:    "valid",
			ui:      NewUserinfo(),
			key:     "key1",
			value:   "val1",
			wantErr: false,
		},
		{
			name:    "key starts with number",
			ui:      NewUserinfo(),
			key:     "1key",
			value:   "val",
			wantErr: true,
		},
		{
			name:    "key oversized",
			ui:      NewUserinfo(),
			key:     "lkjasflkjsflkjsfkljsfdkljsdfkljsfkljsdfkljsdfkljsdflklkjsdflkjsflkjs",
			value:   "val",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ui.Add(tc.key, tc.value)
			if (got != nil) != tc.wantErr {
				t.Error("Unexpected error:", got)
			}
		})
	}
}

func TestUserinfoMarshal(t *testing.T) {
	tests := []struct {
		name string
		ui   Userinfo
		want string
	}{
		{
			name: "valid",
			ui: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
			want: "\\key1\\val1\\key2\\val2",
		},
		{
			name: "empty",
			ui:   map[string]string{},
			want: "",
		},
		{
			name: "out of order",
			ui: map[string]string{
				"key_dog": "val1",
				"key_cat": "val2",
			},
			want: "\\key_cat\\val2\\key_dog\\val1",
		},
		{
			name: "oversized value",
			ui: map[string]string{
				"key_dog": "val1lkajsflkjsflkjasflkasflkjsfkljwfieifjijeflijelfjeflkjelkjlekfjkjlklejlkj",
				"key_cat": "val2",
			},
			want: "\\key_cat\\val2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.ui.Marshal()
			if got != tc.want {
				t.Errorf("Marshal() = %q, want %q", got, tc.want)
			}
		})
	}
}
