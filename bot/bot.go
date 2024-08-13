package bot

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"

	"github.com/packetflinger/libq2/message"
	pl "github.com/packetflinger/libq2/player"
)

const (
	MaxMessageSize = 1390
)

type Bot struct {
	Net     Connection
	User    pl.Userinfo
	Version string
	Netchan NetChan
	Spawned bool
	Debug   bool
}

type Connection struct {
	Address   string
	Port      int
	Conn      net.Conn
	Challenge message.ChallengeResponse
}

type NetChan struct {
	msgIn      message.MessageBuffer
	msgOut     message.MessageBuffer
	QPort      uint16
	Sequence1  int32
	Sequence2  int32
	ReliableS1 bool
	ReliableS2 bool
}

// is there anything that needs to be sent?
func (bot *Bot) OutPending() bool {
	return len(bot.Netchan.msgOut.Buffer) > 0
}

// was a recently received msg reliable and needs an ack?
func (bot *Bot) ReliablePending() bool {
	return bot.Netchan.ReliableS2
}

// Any sequence sent with reliable bit set (0x800000)
func (bot *Bot) SendAck() error {
	err := bot.Send()
	if err != nil {
		return err
	}
	return nil
}

func (bot *Bot) ClientCommand(str string, reliable bool) error {
	msg := message.MessageBuffer{}
	msg.WriteString(str)
	p := message.ClientPacket{
		Sequence1:   bot.Netchan.Sequence1,
		Sequence2:   bot.Netchan.Sequence2,
		QPort:       bot.Netchan.QPort,
		Reliable1:   reliable,
		Reliable2:   bot.Netchan.ReliableS2,
		MessageType: message.CLCStringCommand,
		Data:        msg.Buffer,
	}
	packet := p.Marshal()
	if bot.Debug {
		fmt.Printf("sending\n%s\n", hex.Dump(packet))
	}
	_, e := bot.Net.Conn.Write(packet)
	if e != nil {
		return e
	}
	return nil
}

func (bot *Bot) Run(cb message.Callback) error {
	if bot.Netchan.QPort == 0 {
		bot.Netchan.QPort = uint16(rand.Intn(256))
	}
	c, e := net.Dial("udp4", fmt.Sprintf("%s:%d", bot.Net.Address, bot.Net.Port))
	if e != nil {
		return e
	}
	bot.Net.Conn = c
	bot.Netchan.Sequence1 = 1
	bot.Netchan.Sequence2 = 0
	bot.Netchan.ReliableS1 = true

	defer c.Close()
	log.Println("requesting challenge...")

	getchal := message.ConnectionlessPacket{Data: "getchallenge"}.Marshal()
	_, e = c.Write(getchal)
	if e != nil {
		return e
	}

	chal := make([]byte, 40)
	_, e = c.Read(chal)
	if e != nil {
		return e
	}

	cmsg := message.MessageBuffer{Buffer: chal}
	ch, err := cmsg.ParseChallenge()
	if err != nil {
		return err
	}

	bot.Net.Challenge = ch
	log.Printf("received challenge (%d)\n", bot.Net.Challenge.Number)

	constr := fmt.Sprintf("connect 34 %d %d \"%s\"", bot.Netchan.QPort, bot.Net.Challenge.Number, bot.User.Marshal())
	con := message.ConnectionlessPacket{Data: constr}.Marshal()
	_, e = c.Write(con)
	if e != nil {
		return e
	}
	log.Println("connecting...")

	// client_connect ac=1 dlserver=http://[...] map=q2dm1
	input := make([]byte, 100)
	_, e = c.Read(input)
	if e != nil {
		return e
	}
	if bot.Debug {
		fmt.Printf("%s\n", hex.Dump(input))
	}

	bot.ClientCommand("new", true)

	if cb.OnConnect != nil {
		cb.OnConnect()
	}

	for {
		bytes, err := bot.Receive()
		if err != nil {
			return err
		}
		if bytes == 0 {
			break
		}

		serverframe, err := message.ParseMessageLump(bot.Netchan.msgIn, message.Callback{}, cb)
		if err != nil {
			return err
		}

		for _, st := range serverframe.Stuffs {
			// entering the game
			if t := strings.Fields(st.String); len(t) > 1 && t[0] == "precache" {
				bot.Spawned = true
				log.Println("entering game")
				bot.Netchan.msgOut.WriteByte(message.CLCStringCommand)
				bot.Netchan.msgOut.WriteString("begin " + t[1] + "\n")
				bot.Netchan.ReliableS1 = true

				if cb.OnEnter != nil {
					cb.OnEnter()
				}
			}

			// handle version probe
			if t := strings.Fields(st.String); len(t) >= 4 && t[0] == "cmd" && t[2] == "version" {
				bot.Netchan.msgOut.WriteByte(message.CLCStringCommand)
				bot.Netchan.msgOut.WriteString("\177c version " + bot.Version + "\n")
				bot.Netchan.ReliableS1 = true
			}
		}

		// stuff needs sending
		if bot.OutPending() {
			err = bot.Send()
			if err != nil {
				log.Println(err)
			}
			continue
		} else {
			if bot.ReliablePending() || !bot.Spawned {
				err = bot.SendAck()
				if err != nil {
					log.Println(err)
				}
				continue
			}
		}
		/*
			if err != nil {
				log.Println(err)
			}
		*/
		if bot.Netchan.Sequence2&3 == 0 && bot.Spawned {
			err = bot.SendAck()
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

func (bot *Bot) Send() error {
	msg2 := &bot.Netchan.msgOut
	msg := message.MessageBuffer{}
	msg.WriteLong(bot.Netchan.Sequence1)
	if bot.Netchan.ReliableS1 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteLong(bot.Netchan.Sequence2)
	if bot.Netchan.ReliableS2 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteShort(uint16(bot.Netchan.QPort))

	if len(msg2.Buffer) > 0 {
		msg.Buffer = append(msg.Buffer, msg2.Buffer...)
		msg.Index += msg2.Index
	}

	_, e := bot.Net.Conn.Write(msg.Buffer)
	if e != nil {
		return e
	}

	if bot.Debug {
		fmt.Printf("sent:\n%s\n", hex.Dump(msg.Buffer))
	}

	bot.Netchan.Sequence1++
	bot.Netchan.ReliableS1 = false
	msg2.Reset()
	return nil
}

func (bot *Bot) Receive() (int, error) {
	in := make([]byte, MaxMessageSize*1.5)
	bytes, error := bot.Net.Conn.Read(in)
	if error != nil {
		return bytes, error
	}

	msg := &bot.Netchan.msgIn
	msg.Reset()
	msg.Buffer = in[:bytes]

	if bot.Debug {
		fmt.Printf("received\n%s\n", hex.Dump(msg.Buffer))
	}

	// normally this would be a ReadLong(), but we need it to stay in byte slice format
	// in order to check for reliability flag and strip it if present
	sequence := msg.ReadData(4)

	// is the last bit (sign bit) 1?
	if sequence[3]&0x80 > 0 {
		bot.Netchan.ReliableS2 = true

		// flip it back to 0
		sequence[3] ^= 0x80
	} else {
		bot.Netchan.ReliableS2 = false
	}
	tmpmsg := message.MessageBuffer{Buffer: sequence}
	bot.Netchan.Sequence2 = tmpmsg.ReadLong()

	// we don't care about the ack sequence
	_ = msg.ReadLong()

	return bytes, nil
}
