package state

import (
	"testing"
)

func TestFetchInfo(t *testing.T) {
	server := Server{
		Address: "frag.gr",
		Port:    27910,
	}

	info, err := server.FetchInfo()
	if err != nil {
		t.Error(err)
	}

	if info.serverInfo["hostname"] != "PacketFlinger.com ~ DM" {
		t.Error("invalid hostname lookup")
	}
}
