package state

import (
	"testing"
)

func TestDoRcon(t *testing.T) {
	server := Server{
		Address:  "frag.gr",
		Port:     27910,
		Password: "9E7A365A-A203-4010-A21E-9A9CCFB357D",
	}

	_, err := server.DoRcon("status")
	if err != nil {
		t.Error(err)
	}
}
