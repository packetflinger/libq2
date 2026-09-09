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
	"sync"
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

	// Move is the usercmd sent every frame.  Without it a bot can connect and
	// talk but cannot walk: BuildUserCommand built an empty command and threw
	// lastMove away, so every bot stood still.  Write it from a callback or
	// another goroutine and the next frame carries it.
	Move pl.UserCommand
	// MoveMu guards Move for callers driving from another goroutine.
	MoveMu sync.Mutex
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
	bot.oldframes = make(map[int32]*pb.Frame)

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

			// Remember each frame: the server delta-compresses the next one
			// against the last frame we acked, so without a history of them
			// ParsePacket has nothing to merge against and every value the
			// server left out reads back as zero -- a standing player's origin
			// most of all, since it is omitted precisely when it has not
			// changed.
			for _, fr := range packet.GetFrames() {
				bot.oldframes[fr.GetNumber()] = fr
				for n := range bot.oldframes {
					if fr.GetNumber()-n > 64 {
						delete(bot.oldframes, n)
					}
				}
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
				// EVERY stufftext reaches the callback, including the ones this
				// loop answers itself.  A stufftext is the server typing a console
				// command into this client, and a caller registered on SVCStuffText
				// is asking to see what the server said -- not only the words the
				// bot had no use for.  Dispatching below the handling instead, with
				// every branch returning early, made the channel go silent for
				// exactly the interesting ones: `changing`, `reconnect`, and on a
				// vanilla-protocol server the whole `cmd ...` handshake -- which is
				// every stufftext such a server sends, so the callback received
				// nothing at all for a full session.
				if cb, ok := bot.callbacks[message.SVCStuffText]; ok {
					cb(st, &bot.Netchan.out)
				}

				t := strings.Fields(st.GetData())

				// entering the game
				if len(t) > 1 && t[0] == "precache" {
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

				// A LEVEL CHANGE IS NOT A DISCONNECT, and the two commands
				// that carry one have to be answered or the client goes
				// quiet without ever being dropped.
				//
				// `changing` is the server saying it is leaving this level.
				// id's CL_Changing_f takes the client out of the spawned
				// state and holds the connection open; there is nothing to
				// send back.  Without this the bot keeps sending usercmds
				// for a level that no longer exists, and the fallback at the
				// bottom of this loop echoes the word back as a client
				// command -- which is what a server logs as `<name>:
				// changing`.
				if len(t) >= 1 && t[0] == "changing" {
					bot.Spawned = false
					bot.AckPending = true
					continue
				}

				// `reconnect` asks for the SPAWN handshake again, not for a
				// new connection.  id's CL_Reconnect_f answers a connected
				// client with the string command `new`, and that is what
				// makes the server re-send serverdata, configstrings and
				// baselines for the new level.  A client that does not
				// answer stays connected and receives nothing further: its
				// frame counter stops, its configstrings go stale and every
				// command it sends is for a level the server has left.  From
				// the outside that is indistinguishable from a mod that has
				// stopped talking to it, which is how it was first read.
				//
				// The frame history goes with it.  Frames are delta-encoded
				// against earlier ones and the new level restarts the
				// sequence, so keeping the old map's frames as a delta base
				// decodes the new level against the wrong entities.
				if len(t) >= 1 && t[0] == "reconnect" {
					bot.Spawned = false
					bot.FrameNum = 0
					clear(bot.oldframes)
					bot.AddClientString("new\n")
					bot.Netchan.ReliableS1 = true
					bot.AckPending = true
					continue
				}

				// handle version probe
				if len(t) >= 4 && t[0] == "cmd" && t[2] == "version" {
					bot.AddClientString("\177c version %s\n", bot.Version)
					bot.Netchan.ReliableS1 = true
					continue
				}

				// `cmd <rest>` MEANS "forward <rest> to the server", and a
				// vanilla-protocol server's whole spawn handshake is built out
				// of it: SV_New_f stuffs `cmd configstrings <spawncount> 0`,
				// SV_Configstrings_f answers with more of the same and then
				// `cmd baselines <spawncount> 0`, and only after the baselines
				// are drained does `precache <spawncount>` arrive.  Q2PRO
				// short-circuits all of that and stuffs `precache` straight
				// away, which is why this was never needed before -- against
				// Yamagi Quake II, id's own q2ded or r1q2 the client stalled
				// here for ever, and the fallback below sent the text back
				// WITH its `cmd` prefix, which the server logs as an unknown
				// client command.
				//
				// The text is macro-expanded on the way out.  A server that asks
				// for a cvar value -- `cmd \177c <var> $<var>`, which is how q2pro
				// serves a cvarban and how it collects an anticheat token -- gets
				// an answer, and a literal `$<var>` is not an answer: it is a value
				// the client claims to hold, and a ban rule matches or misses on it.
				if len(t) >= 2 && t[0] == "cmd" {
					bot.AddClientString("%s\n", strings.Join(bot.expandCVars(t[1:]), " "))
					bot.Netchan.ReliableS1 = true
					bot.AckPending = true
					continue
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

			// Layouts and centerprints are parsed out of the packet but were
			// never handed to a callback, so a caller could register for them
			// and never hear anything.  They are how a mod talks to one
			// client: the scoreboard `score` draws, the round and match
			// announcements a team mod centers on screen.
			for _, l := range packet.GetLayouts() {
				if cb, ok := bot.callbacks[message.SVCLayout]; ok {
					cb(l, &bot.Netchan.out)
				}
			}
			for _, cp := range packet.GetCenterprints() {
				if cb, ok := bot.callbacks[message.SVCCenterPrint]; ok {
					cb(cp, &bot.Netchan.out)
				}
			}
			for _, b := range packet.GetBaselines() {
				cb, ok := bot.callbacks[message.SVCSpawnBaseline]
				if ok {
					cb(b, &bot.Netchan.out)
				}
			}

			// Sounds were parsed into the packet and dropped on the floor, the
			// same way layouts and centerprints were.  A sound is the only
			// evidence some mod behaviour leaves: a death scream, a pickup, an
			// announcer cue.  Handed to a callback keyed on SVCSound, which
			// receives *pb.PackedSound and can map Index through the
			// CS_SOUNDS block to a name.
			for _, snd := range packet.GetSounds() {
				if cb, ok := bot.callbacks[message.SVCSound]; ok {
					cb(snd, &bot.Netchan.out)
				}
			}

			// Only once spawned: between `changing` and the `begin` that
			// follows `precache` there is no level to move in, and what has to
			// get through is the reliable `new`.
			if bot.Spawned {
				bot.Netchan.out.Append(bot.BuildUserCommand())
			}
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

// SendUserinfo pushes the bot's current User map to the server as a userinfo
// update, which is how a real client tells the server its name or skin changed.
// Without it the map could be edited but never sent, so a mod's
// ClientUserinfoChanged path was unreachable from a bot.
func (b *Bot) SendUserinfo() {
	b.Netchan.out.Append(ClientUserMessage(b.User.Marshal()))
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
	b.MoveMu.Lock()
	move := b.Move
	b.MoveMu.Unlock()
	move.LightLevel = 150
	if move.Msec == 0 {
		move.Msec = 100
	}

	// The checksummed region is everything AFTER the checksum byte: the
	// acknowledged frame number and the three commands.  Built first so the
	// byte can be computed over it rather than invented.
	//
	// Three commands per packet is what the protocol expects: the oldest two
	// are re-sends so a dropped packet does not lose input.
	body := message.NewEmptyBuffer()
	body.WriteLong(b.FrameNum)
	body.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))
	body.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))
	body.Append(move.WriteDeltaUsercmd(pl.UserCommand{}))

	// A REAL CHECKSUM, not a placeholder.  Vanilla-protocol servers verify it
	// and silently ignore the rest of the packet when it is wrong, so with a
	// made-up byte a client connects, spawns and then never moves on Yamagi
	// Quake II, q2ded or r1q2.  Q2PRO does not check, which is why the
	// placeholder went unnoticed.  The sequence it is salted with is the
	// OUTGOING one this packet will carry.
	msg := message.NewEmptyBuffer()
	msg.WriteByte(message.CLCMove)
	msg.WriteByte(int(blockSequenceCRCByte(body.Data, b.Netchan.Sequence1)))
	msg.Append(body)
	return msg
}

// cvar returns the bot's value for a cvar and whether it had one, matched
// case-insensitively the way Quake II's own cvar lookup is.
func (b *Bot) cvar(name string) (string, bool) {
	for k, v := range b.CVars {
		if strings.EqualFold(name, k) {
			return v, true
		}
	}
	return "", false
}

// expandCVars replaces every $name in tokens with the bot's value for that
// cvar, and with the empty string when it has none -- which is what a real
// client's macro expansion does, since an unset cvar expands to nothing.
//
// It exists apart from ResolveString because the two answer different
// questions.  ResolveString prepares text for this bot's own command table,
// resolves an alias in the first token, and DROPS a $name it cannot resolve --
// which shifts every argument after it one place left.  Text being forwarded
// to the server must keep its shape: the server counts arguments, so an
// unknown value has to stay an empty argument rather than vanish.
func (b *Bot) expandCVars(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if len(t) < 2 || !strings.HasPrefix(t, "$") {
			out = append(out, t)
			continue
		}
		val, _ := b.cvar(t[1:])
		out = append(out, val)
	}
	return out
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
		if v, ok := b.cvar(t[1:]); ok {
			out = append(out, v)
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
