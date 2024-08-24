package pak

import (
	"os"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	data, err := os.ReadFile("../testdata/test.pak")
	if err != nil {
		t.Error(err)
	}

	archive, err := Unmarshal(data)
	if err != nil {
		t.Error(err)
	}
	t.Error(archive)
}

func TestMarshal(t *testing.T) {
	data, err := os.ReadFile("../testdata/test.pak")
	if err != nil {
		t.Error(err)
	}

	archive, err := Unmarshal(data)
	if err != nil {
		t.Error(err)
	}

	data, err = Marshal(archive)
	if err != nil {
		t.Error(err)
	}
	err = os.WriteFile("../testdata/test-out.pak", data, 0777)
	if err != nil {
		t.Error(err)
	}
}
