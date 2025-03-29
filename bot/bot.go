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

	pb "github.com/packetflinger/libq2/proto"
)

const (
	MaxMessageSize = 1390
)

type Bot struct {
	Net       Connection
	User      pl.Userinfo
	Version   string
	Netchan   NetChan
	Spawned   bool
	Debug     bool
	callbacks map[int]func(any, *message.Buffer)
	oldframes map[int32]*pb.Frame
}

type Connection struct {
	Address   string
	Port      int
	Conn      net.Conn
	Challenge *pb.Challenge
}

type NetChan struct {
	in         message.Buffer
	out        message.Buffer
	QPort      uint16
	Sequence1  int
	Sequence2  int
	ReliableS1 bool
	ReliableS2 bool
}

func (b *Bot) RegisterCallback(index int, dofunc func(any, *message.Buffer)) {
	if b.callbacks == nil {
		b.callbacks = make(map[int]func(any, *message.Buffer))
	}
	b.callbacks[index] = dofunc
}

func (b *Bot) UnregisterCallback(index int) {
	if b.callbacks == nil {
		b.callbacks = make(map[int]func(any, *message.Buffer))
	}
	delete(b.callbacks, index)
}

// is there anything that needs to be sent?
func (bot *Bot) OutPending() bool {
	return len(bot.Netchan.out.Data) > 0
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
	msg := message.Buffer{}
	msg.WriteString(str)
	p := message.ClientPacket{
		Sequence1:   bot.Netchan.Sequence1,
		Sequence2:   bot.Netchan.Sequence2,
		QPort:       bot.Netchan.QPort,
		Reliable1:   reliable,
		Reliable2:   bot.Netchan.ReliableS2,
		MessageType: message.CLCStringCommand,
		Data:        msg.Data,
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

func (bot *Bot) Run() error {
	if bot.Netchan.QPort == 0 {
		bot.Netchan.QPort = uint16(rand.Intn(256))
	}
	addr := fmt.Sprintf("%s:%d", bot.Net.Address, bot.Net.Port)
	c, e := net.Dial("udp4", addr)
	if e != nil {
		return e
	}
	bot.Net.Conn = c
	bot.Netchan.Sequence1 = 1
	bot.Netchan.Sequence2 = 0
	bot.Netchan.ReliableS1 = true

	defer c.Close()
	log.Println("requesting challenge from", addr)

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

	cmsg := message.Buffer{Data: chal}
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

	for {
		bytes, err := bot.Receive()
		if err != nil {
			return err
		}
		if bytes == 0 {
			break
		}

		packet, err := bot.Netchan.in.ParsePacket(bot.oldframes)
		if err != nil {
			return err
		}
		for _, pr := range packet.GetPrints() {
			log.Printf("%s", pr.GetData()) // newline is included in the string
			cb, ok := bot.callbacks[message.SVCPrint]
			if ok {
				cb(pr, &bot.Netchan.out)
			}
		}

		for _, st := range packet.GetStuffs() {
			// entering the game
			if t := strings.Fields(st.GetData()); len(t) > 1 && t[0] == "precache" {
				bot.Spawned = true
				log.Println("entering game")
				bot.Netchan.out.WriteByte(message.CLCStringCommand)
				bot.Netchan.out.WriteString("begin " + t[1] + "\n")
				bot.Netchan.ReliableS1 = true
				cb, ok := bot.callbacks[message.CallbackOnBegin]
				if ok {
					cb(nil, &bot.Netchan.out)
				}
				continue
			}

			// handle version probe
			if t := strings.Fields(st.GetData()); len(t) >= 4 && t[0] == "cmd" && t[2] == "version" {
				bot.Netchan.out.WriteByte(message.CLCStringCommand)
				bot.Netchan.out.WriteString("\177c version " + bot.Version + "\n")
				bot.Netchan.ReliableS1 = true
			}
			cb, ok := bot.callbacks[message.SVCStuffText]
			if ok {
				cb(st, &bot.Netchan.out)
			}
		}

		for _, frame := range packet.GetFrames() {
			cb, ok := bot.callbacks[message.SVCFrame]
			if ok {
				cb(frame, &bot.Netchan.out)
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
		if ((bot.Netchan.Sequence2 & 3) == 0) && bot.Spawned {
			err = bot.SendAck()
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

func (bot *Bot) Send() error {
	msg2 := &bot.Netchan.out
	msg := message.Buffer{}
	msg.WriteLong(bot.Netchan.Sequence1)
	if bot.Netchan.ReliableS1 {
		msg.Data[msg.Index-1] |= 0x80
	}
	msg.WriteLong(bot.Netchan.Sequence2)
	if bot.Netchan.ReliableS2 {
		msg.Data[msg.Index-1] |= 0x80
	}
	msg.WriteShort(int(bot.Netchan.QPort))

	if len(msg2.Data) > 0 {
		msg.Data = append(msg.Data, msg2.Data...)
		msg.Index += msg2.Index
	}

	_, e := bot.Net.Conn.Write(msg.Data)
	if e != nil {
		return e
	}

	if bot.Debug {
		fmt.Printf("sent:\n%s\n", hex.Dump(msg.Data))
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

	msg := &bot.Netchan.in
	msg.Reset()
	msg.Data = in[:bytes]

	if bot.Debug {
		fmt.Printf("received\n%s\n", hex.Dump(msg.Data))
	}

	// normally this would be a ReadLong(), but we need it to stay in byte slice format
	// in order to check for reliability flag and strip it if present
	sequence := msg.ReadData(4)

	// is the last bit (sign bit) 1?
	if (sequence[3] & 0x80) > 0 {
		bot.Netchan.ReliableS2 = true

		// flip it back to 0
		sequence[3] ^= 0x80
	} else {
		bot.Netchan.ReliableS2 = false
	}
	tmpmsg := message.Buffer{Data: sequence}
	bot.Netchan.Sequence2 = tmpmsg.ReadLong()

	// we don't care about the ack sequence
	_ = msg.ReadLong()

	return bytes, nil
}
