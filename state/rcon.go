package state

import (
	"fmt"

	"github.com/packetflinger/libq2/message"
)

type Rcon struct {
	Serverinfo *Server
	Input      string
	Output     string
}

func (s *Server) DoRcon(str string) (Rcon, error) {
	rcon := Rcon{
		Serverinfo: s,
		Input:      str,
	}
	p := message.ConnectionlessPacket{
		Data: fmt.Sprintf("rcon %s %s\n", s.Password, str),
	}
	out, err := p.Send(s.Address, s.Port)
	if err != nil {
		return Rcon{}, err
	}

	rcon.Output = string(out.Buffer[10:])
	return rcon, nil
}
