// A very basic library for making a bot capable of connecting to a Quake 2
// server.
package bot

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/packetflinger/libq2/message"

	pl "github.com/packetflinger/libq2/player"
	pb "github.com/packetflinger/libq2/proto"
)

const (
	MoveMask       = 1 << 4
	MaxMessageSize = 1390
	LightLevel     = 150
)

var (
	hz        = 10
	frametime = float64(1000.0 / hz)
)

var (
	commands = map[string]func(*Bot, Cmd){
		"alias": aliasFunc,
		"exec":  nullFunc,
		"quit":  quitFunc,
		"say":   sayFunc,
		"set":   setFunc,
	}
)

type Bot struct {
	Net        Connection
	User       pl.Userinfo
	Version    string
	Netchan    NetChan
	Spawned    bool
	AckPending bool
	Debug      bool
	callbacks  map[int]func(any, *message.Buffer)
	oldframes  map[int32]*pb.Frame
	lastMove   pl.UserCommand // usercmd_t
	FrameNum   int
	oldMoves   [MoveMask]pl.UserCommand
	Aliases    map[string]string
	CVars      map[string]string
	Cmds       map[string]func(*Bot, Cmd)
}

type Connection struct {
	Address   string
	Port      int
	Conn      net.Conn
	Challenge *pb.Challenge
}

type NetChan struct {
	in           message.Buffer
	out          message.Buffer
	QPort        int
	Sequence1    int
	Sequence2    int
	ReliableS1   bool
	ReliableS2   bool
	LastReliable int
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
	bot.Cmds = commands
	recv := make(chan bool)
	stop := make(chan bool)

	if bot.Netchan.QPort == 0 {
		bot.Netchan.QPort = rand.Intn(256)
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

	cmsg := message.NewBuffer(chal)
	ch, err := cmsg.ParseChallenge()
	if err != nil {
		return err
	}

	bot.Net.Challenge = ch
	log.Printf("received challenge [%d]\n", bot.Net.Challenge.Number)

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

	go func() {
		for {
			bytes, err := bot.Receive()
			if err != nil {
				return
			}
			if bytes == 0 {
				stop <- true
				break
			}
			// just sequence and ack sequence, ack back
			if bytes == 8 {
				bot.AckPending = true
			}
			recv <- true

			packet, err := bot.Netchan.in.ParsePacket(bot.oldframes)
			if err != nil {
				return
			}

			for _, fr := range packet.GetFrames() {
				bot.FrameNum = int(fr.GetNumber())
				cb, ok := bot.callbacks[message.SVCFrame]
				if ok {
					cb(fr, &bot.Netchan.out)
				}
			}

			for _, pr := range packet.GetPrints() {
				cb, ok := bot.callbacks[message.SVCPrint]
				if ok {
					cb(pr, &bot.Netchan.out)
				}
			}

			for _, st := range packet.GetStuffs() {
				// entering the game
				if t := strings.Fields(st.GetData()); len(t) > 1 && t[0] == "precache" {
					bot.Spawned = true
					log.Println("spawning into game")
					bot.AddClientString("begin %s\n", t[1])
					bot.Netchan.ReliableS1 = true
					bot.FrameNum = 1
					cb, ok := bot.callbacks[message.CallbackOnBegin]
					if ok {
						cb(nil, &bot.Netchan.out)
					}
					continue
				}

				// handle version probe
				if t := strings.Fields(st.GetData()); len(t) >= 4 && t[0] == "cmd" && t[2] == "version" {
					bot.AddClientString("\177c version %s\n", bot.Version)
					bot.Netchan.ReliableS1 = true
					continue
				}

				if cb, ok := bot.callbacks[message.SVCStuffText]; ok {
					cb(st, &bot.Netchan.out)
				}
				resolved := bot.ResolveString(st.GetData())
				cmds := ParseCmd(resolved)
				for _, c := range cmds {
					if cmd, found := bot.Cmds[c.commandName]; found {
						cmd(bot, c)
					} else {
						sayFunc(bot, c)
					}
				}
				bot.AckPending = true
			}

			for _, cs := range packet.GetConfigStrings() {
				cb, ok := bot.callbacks[message.SVCConfigString]
				if ok {
					cb(cs, &bot.Netchan.out)
				}
			}
			for _, b := range packet.GetBaselines() {
				cb, ok := bot.callbacks[message.SVCSpawnBaseline]
				if ok {
					cb(b, &bot.Netchan.out)
				}
			}
			for _, frame := range packet.GetFrames() {
				cb, ok := bot.callbacks[message.SVCFrame]
				if ok {
					cb(frame, &bot.Netchan.out)
				}
			}

			bot.Netchan.out.Append(bot.BuildUserCommand())
			bot.Send()
		}
	}()

	var usercmd pl.UserCommand
	for {
		select {
		case _ = <-recv:
		//fmt.Println("recv'd something")
		case <-stop:
			fmt.Println("instructed to quit")
			return nil
		case <-time.After(time.Duration(frametime) * time.Millisecond):
			if bot.Spawned {
				usercmd = pl.UserCommand{
					Msec: 100,
				}
				bot.Netchan.out.Append(bot.BuildUserCommand())
				bot.lastMove = usercmd
				bot.Send()
			} else {
				if !bot.Netchan.out.IsEmpty() || bot.AckPending {
					bot.Send()
				}
			}
		}
	}
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
	bot.AckPending = false
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
	msg.Length = len(msg.Data)

	if bot.Debug {
		fmt.Printf("received\n%s\n", hex.Dump(msg.Data))
	}

	sequence := uint32(msg.ReadLong())
	reliable := (sequence >> 31) == 1
	bot.Netchan.Sequence2 = int(sequence) & ^(1 >> 31)
	if reliable {
		bot.Netchan.ReliableS2 = true
		bot.Send() // immediately ack if last is reliable
	} else {
		bot.Netchan.ReliableS2 = false
	}

	// we don't care about the ack sequence
	_ = msg.ReadLong()

	return bytes, nil
}

// Marshal a c2s userinfo update message
func ClientUserMessage(ui string) message.Buffer {
	msg := message.NewEmptyBuffer()
	msg.WriteByte(message.CLCUserinfo)
	msg.WriteString(ui) // maybe use WriteData?
	return msg
}

func ClientStringCommand(s string) message.Buffer {
	msg := message.NewEmptyBuffer()
	msg.WriteByte(message.CLCStringCommand)
	msg.WriteString(s)
	return msg
}

func (b *Bot) BuildUserCommand() message.Buffer {
	msg := message.NewEmptyBuffer()
	msg.WriteByte(message.CLCMove)
	msg.WriteByte(0xa1) // checksum, make up something
	msg.WriteLong(b.FrameNum)
	move := pl.UserCommand{
		LightLevel: 150,
	}
	msg.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))
	msg.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))
	move.Msec = 100
	msg.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))
	return msg
}

// Replace any variables and aliases with their substitutions. Aliases are not
// recursive.
func (b *Bot) ResolveString(s string) string {
	var out []string
	var alias string
	tokens := strings.Fields(s)
	if len(tokens) == 0 {
		return ""
	}
	for k, v := range b.Aliases {
		if strings.EqualFold(k, tokens[0]) {
			alias = v
		}
	}
	if len(alias) > 0 {
		out = append(out, alias)
	} else {
		out = append(out, tokens[0])
	}
	for _, t := range tokens[1:] {
		// variables start with $, if not skip it
		if !strings.HasPrefix(t, "$") {
			out = append(out, t)
			continue
		}
		for k, v := range b.CVars {
			if strings.EqualFold(t[1:], k) {
				out = append(out, v)
			}
		}
	}
	return strings.Join(out, " ")
}

// Add a client string to the bot's outgoing message buffer. This is how
// strings are sent from the bot to the server.
func (b *Bot) AddClientString(format string, args ...any) {
	final := fmt.Sprintf(format, args...)
	b.Netchan.out.WriteByte(message.CLCStringCommand)
	b.Netchan.out.WriteString(final)
}

// Empty function to associate with commands we want to ignore
func nullFunc(_ *Bot, c Cmd) {
	fmt.Printf("silently dropping command %q\n", c.GetFullCommand())
}

func setFunc(b *Bot, c Cmd) {
	if b.CVars == nil {
		b.CVars = make(map[string]string)
	}
	b.CVars[c.Argv(0)] = c.Argv(1)
}

func aliasFunc(b *Bot, c Cmd) {
	if b.Aliases == nil {
		b.Aliases = make(map[string]string)
	}
	b.Aliases[c.Argv(0)] = c.Argv(1)
}

// Called in the even a "say [...]" command is received, or if an unrecognized
// command is received.
func sayFunc(b *Bot, c Cmd) {
	var what string
	// If it's an unrecognized command, add it to the list of args. Since it
	// has to be first though, reverse the args, append, and reverse again.
	if c.GetCommand() != "say" {
		slices.Reverse(c.arguments)
		c.arguments = append(c.arguments, c.GetCommand())
		slices.Reverse(c.arguments)
	}
	what = strings.Join(c.arguments, " ")
	b.AddClientString(what)
}

func quitFunc(b *Bot, c Cmd) {
}
