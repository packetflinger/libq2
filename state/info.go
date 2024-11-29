package state

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/packetflinger/libq2/message"
)

type ServerInfo struct {
	Server  map[string]string
	Players []struct {
		Name  string
		Score int
		Ping  int
	}
}

func (s *Server) FetchInfo() (ServerInfo, error) {
	p := message.ConnectionlessPacket{
		Data: "status",
	}
	out, err := p.Send(s.Address, s.Port)
	if err != nil {
		return ServerInfo{}, err
	}
	if out.Length < 5 {
		return ServerInfo{}, fmt.Errorf("invalid serverinfo response, server running?")
	}
	lines := strings.Split(strings.Trim(string(out.Buffer[4:]), " \n\t"), "\n")
	return parseServerinfo(lines)
}

func parseServerinfo(s []string) (ServerInfo, error) {
	si := ServerInfo{}
	info := map[string]string{}

	if s[0] != "print" {
		return ServerInfo{}, fmt.Errorf("invalid server info string")
	}

	serverinfo := ""
	if len(s) > 1 {
		if len(s[1]) > 0 {
			serverinfo = s[1][1:]
			vars := strings.Split(serverinfo, "\\")

			for i := 0; i < len(vars); i += 2 {
				info[strings.ToLower(vars[i])] = vars[i+1]
			}
		}
	}

	if len(s) > 2 {
		players := s[2:]
		info["player_count"] = fmt.Sprintf("%d", len(players))
		if len(players) > 0 {
			playernames := ""

			for _, p := range players {
				player := strings.SplitN(p, " ", 3)
				playernames = fmt.Sprintf("%s,%s", playernames, player[2])
				score, _ := strconv.Atoi(player[0])
				ping, _ := strconv.Atoi(player[1])
				si.Players = append(si.Players, struct {
					Name  string
					Score int
					Ping  int
				}{
					Name:  strings.Trim(player[2], "\""), // take quotes off
					Score: score,
					Ping:  ping,
				})
			}

			info["players"] = playernames[1:]
		}
	} else {
		info["player_count"] = "0"
	}

	si.Server = info
	return si, nil
}

// Parse and info string (in the format of:
// "\key1\val1\key2\val2\key3\val3\" into a key/value map)
func ParseInfoString(info string) map[string]string {
	infomap := map[string]string{}
	vars := strings.Split(info, "\\")

	for i := 0; i < len(vars); i += 2 {
		infomap[strings.ToLower(vars[i])] = vars[i+1]
	}
	return infomap
}
