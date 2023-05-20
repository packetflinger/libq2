package bot

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	m "github.com/packetflinger/libq2/message"
	pl "github.com/packetflinger/libq2/player"
)

const (
	MaxMessageSize = 1390
)

type Bot struct {
	Net  Connection
	User pl.Userinfo
}

type Connection struct {
	Address   string
	Port      int
	Conn      net.Conn
	Challenge m.ChallengeResponse
}

func (bot Bot) Run() error {
	c, e := net.Dial("udp4", fmt.Sprintf("%s:%d", bot.Net.Address, bot.Net.Port))
	if e != nil {
		return e
	}
	defer c.Close()
	log.Println("requesting challenge...")
	msg := m.NewMessageBuffer(make([]byte, 17))
	msg.WriteLong(-1)
	msg.WriteString("getchallenge")
	_, e = c.Write(msg.Buffer)
	if e != nil {
		return e
	}

	chal := make([]byte, 40)
	_, e = c.Read(chal)
	if e != nil {
		return e
	}

	cmsg := m.MessageBuffer{Buffer: chal}
	fmt.Printf("%s\n", hex.Dump(cmsg.Buffer))
	ch, err := cmsg.ParseChallenge()
	if err != nil {
		return err
	}

	bot.Net.Challenge = ch
	log.Println(bot.Net.Challenge)
	return nil
}
