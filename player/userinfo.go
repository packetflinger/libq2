package player

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const (
	UserinfoMaxSize      = 512
	UserinfoMaxKeySize   = 64
	UserinfoMaxValueSize = 64
)

var (
	ErrUiBlank         = errors.New("userinfo string is empty")
	ErrUiKeyBegins     = errors.New("key doesn't begin with a letter")
	ErrUiKeyBlank      = errors.New("key string is empty")
	ErrUiKeyNotFound   = errors.New("key not found in userinfo")
	ErrUiKeyOverflow   = errors.New("key string is oversized")
	ErrUiMalformed     = errors.New("userinfo string is malformed")
	ErrUiMismatch      = errors.New("key/value mismatch")
	ErrUiOverflow      = errors.New("userinfo string is oversized")
	ErrUiValueBlank    = errors.New("value string is empty")
	ErrUiValueOverflow = errors.New("value string is oversized")
)

// Use a map since userinfo is basically a key/value store and any key can be
// set by client (or stuffed by the server)
type Userinfo map[string]string

// Allocate memory for the uerinfo map
func NewUserinfo() Userinfo {
	return make(Userinfo)
}

// Add a key/value pair to the userinfo. If they key already exists, the value
// is updated
func (ui Userinfo) Add(k, v string) error {
	if len(k) > UserinfoMaxKeySize {
		return fmt.Errorf("key length oversized: %d > %d", len(k), UserinfoMaxKeySize)
	}
	if len(v) > UserinfoMaxValueSize {
		return fmt.Errorf("value length oversized: %d > %d", len(v), UserinfoMaxValueSize)
	}
	if k == "" {
		return ErrUiKeyBlank
	}
	if v == "" {
		return ErrUiValueBlank
	}
	letter, err := regexp.MatchString(`^[a-zA-Z].*$`, k)
	if err != nil {
		return fmt.Errorf("matching key begins with letter")
	}
	if !letter {
		return ErrUiKeyBegins
	}
	ui[k] = v
	return nil
}

// Get a value from the userinfo
func (ui Userinfo) Value(k string) (string, error) {
	v, ok := ui[k]
	if !ok {
		return "", fmt.Errorf("key not found %q", k)
	}
	return v, nil
}

// Remove a keypair from the userinfo
func (ui Userinfo) Remove(k string) (bool, error) {
	if k == "" {
		return false, ErrUiKeyBlank
	}
	_, ok := ui[k]
	if !ok {
		return false, fmt.Errorf("key not found %q", k)
	}
	delete(ui, k)
	return true, nil
}

// Parse a ui string into a map
func Unmarshal(s string) (Userinfo, error) {
	if s == "" {
		return nil, ErrUiBlank
	}
	if !strings.HasPrefix(s, "\\") {
		return nil, ErrUiMalformed
	}
	if len(s) > UserinfoMaxSize {
		return nil, ErrUiOverflow
	}
	tokens := strings.Split(s[1:], "\\")
	if len(tokens)%2 != 0 {
		return nil, ErrUiMismatch
	}
	ui := NewUserinfo()
	for i := 0; i < len(tokens); i++ {
		k, v := tokens[i], tokens[i+1]
		if len(k) > UserinfoMaxKeySize {
			i++
			continue
		}
		if len(v) > UserinfoMaxValueSize {
			i++
			continue
		}
		ui[k] = v
	}
	return ui, nil
}

// Build a userinfo string from a map of userinfo variables. If a key or value
// is oversize, the pair is dropped. If a pair overflows the overall size it's
// dropped. Pairs are sorted alphabetically prior so the output is
// deterministic.
func (ui Userinfo) Marshal() string {
	if ui == nil {
		return ""
	}
	var keys []string
	for k := range ui {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var out, newpair string
	for _, k := range keys {
		if len(k) > UserinfoMaxKeySize {
			continue
		}
		v := ui[k]
		if len(v) > UserinfoMaxValueSize {
			continue
		}
		newpair = fmt.Sprintf("\\%s\\%s", k, v)
		if (len(out) + len(newpair)) > UserinfoMaxSize {
			continue
		}
		out += newpair
	}
	return out
}
