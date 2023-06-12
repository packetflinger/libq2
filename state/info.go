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

	lines := strings.Split(strings.Trim(string(out.Buffer[4:]), " \n\t"), "\n")
	info := parseServerinfo(lines)
	return info, nil
}

func parseServerinfo(s []string) ServerInfo {
	si := ServerInfo{}
	serverinfo := s[1][1:]
	playerinfo := s[2 : len(s)-1]

	info := map[string]string{}
	vars := strings.Split(serverinfo, "\\")

	for i := 0; i < len(vars); i += 2 {
		info[strings.ToLower(vars[i])] = vars[i+1]
	}

	playercount := len(playerinfo)
	info["player_count"] = fmt.Sprintf("%d", playercount)

	if playercount > 0 {
		players := ""

		for _, p := range playerinfo {
			player := strings.SplitN(p, " ", 3)
			players = fmt.Sprintf("%s,%s", players, player[2])
			score, _ := strconv.Atoi(player[0])
			ping, _ := strconv.Atoi(player[1])
			si.Players = append(si.Players, struct {
				Name  string
				Score int
				Ping  int
			}{
				Name:  player[2][1 : len(player[2])-1],
				Score: score,
				Ping:  ping,
			})
		}

		info["players"] = players[1:]
	}
	si.Server = info
	return si
}
