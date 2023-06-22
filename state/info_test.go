package state

import (
	"fmt"
	"testing"
)

func TestFetchInfo(t *testing.T) {
	server := Server{
		Address: "localhost",
		Port:    27910,
	}

	info, err := server.FetchInfo()
	if err != nil {
		t.Error(err)
	}

	/*
		if info.Server["hostname"] != "PacketFlinger.com ~ DM" {
			t.Error("invalid hostname lookup")
		}
	*/

	fmt.Println(info)
	t.Error()
}
