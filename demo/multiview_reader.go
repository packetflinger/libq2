package demo

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/packetflinger/libq2/message"
	"google.golang.org/protobuf/proto"

	pb "github.com/packetflinger/libq2/proto"
)

// This struct is the receiver used for parsing demo data.
type MVD2Parser struct {
	allDemos       []*pb.MvdDemo     // can be multiple per .mvd2 file
	demo           *pb.MvdDemo       // the current one
	index          int               // which demo?
	binaryData     []byte            // .mvd2 file contents
	binaryPosition int               // where in those contents we are
	callbacks      map[int]func(any) // index is svc_msg type
	debug          bool              // print stuff while parsing
	zipped         bool              // file is gzipped
}

// Read the contents of the multi-view demo file and setup a receiver for
// parsing the data into textproto. The data is checked to make sure it's a
// valid demo file and the internal pointer set to just after the magic value.
//
// Multi-view demo files can be deflated using GZIP (rfc1952) compression on
// the fly. Compression is detected using the file header and automatically
// uncompressed if utilized.
func NewMVD2Parser(f string) (*MVD2Parser, error) {
	var gzip bool
	if f == "" {
		return nil, fmt.Errorf("no file specified")
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	msg := message.NewBuffer(data[:4])
	magic := msg.ReadWord()

	if magic == GZIPMagic {
		data, err = ReadGZIPFile(f)
		if err != nil {
			return nil, err
		}
		msg = message.NewBuffer(data[:4])
		gzip = true
	} else {
		msg.Rewind()
	}

	magic = msg.ReadLong()
	if magic != MVDMagic {
		return nil, fmt.Errorf("%q invalid multi-view demo", f)
	}

	return &MVD2Parser{
		binaryData:     data,
		binaryPosition: 4,
		allDemos:       []*pb.MvdDemo{},
		demo:           &pb.MvdDemo{},
		index:          -1,
		zipped:         gzip,
	}, nil
}

// Get the decompressed demo data
func ReadGZIPFile(filename string) ([]byte, error) {
	var out []byte
	file, err := os.Open(filename)
	if err != nil {
		return out, fmt.Errorf("ReadGZIPFile: %v", err)
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return out, fmt.Errorf("error creating gzip reader: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return out, fmt.Errorf("error reading decompressed data: %v", err)
	}
	return content, nil
}

// Load the binary demo into protobuf
func (p *MVD2Parser) Unmarshal() (*pb.MvdDemo, error) {
	demo := &pb.MvdDemo{}
	for {
		data, err := p.NextPacket()
		if err != nil {
			return nil, err
		}
		if data.Length == 0 {
			break
		}
		packet, err := p.ParsePacket(&data)
		if err != nil {
			return nil, err
		}
		demo.Packets = append(demo.Packets, packet)
	}
	return demo, nil
}

// Demos are organized by shards of data that are essentially packets. Even
// though all the data is already known, each shard represents a server packet's
// worth of game data. Each packet is prefixed with a 16 bit integer of the
// size of the packet and then a bunch of individual messages in any particular
// order. Regular DM2 demos use 4 bytes for this value instead of 2.
//
// There is no ending delimiter to the packet, the end of the last message is
// right up against the size value for the next packet. When the size value is
// 0 (0x0000), there are no more packets and the end of the demo has been
// reached. This is different from regular DM2 demos where a value of -1
// (signed 0xffff) is the end of the demo.
func (p *MVD2Parser) NextPacket() (message.Buffer, error) {
	if p.binaryPosition >= len(p.binaryData) {
		return message.Buffer{}, fmt.Errorf("read past end of packet")
	}
	sizebytes := message.NewBuffer(p.binaryData[p.binaryPosition : p.binaryPosition+2])
	packetLen := int(sizebytes.ReadShort())
	if packetLen == 0 { // EoD
		return message.Buffer{}, nil
	}
	p.binaryPosition += 2
	packet := message.NewBuffer(p.binaryData[p.binaryPosition : p.binaryPosition+packetLen])
	p.binaryPosition += packetLen
	return packet, nil
}

// Loop through all the messages included in this shard of data.
func (p *MVD2Parser) ParsePacket(msg *message.Buffer) (*pb.MvdPacket, error) {
	packet := &pb.MvdPacket{}
	if msg == nil {
		return nil, fmt.Errorf("nil msg buffer")
	}
	for {
		if msg.Index >= msg.Length {
			break
		}

		// Additional bits can be multiplexed with the command. The last 3 bits
		// can be used to hold extra data
		cmd := msg.ReadByte()
		extra := cmd >> CommandBits
		cmd &= CommandMask

		switch cmd {
		case MVDSvcServerData:
			err := p.ParseServerData(msg, extra)
			if err != nil {
				return nil, err
			}
			p.demo.Configstrings = p.ParseConfigStrings(msg)
			frame, err := p.ParseFrame(msg)
			if err != nil {
				return nil, err
			}
			p.demo.Entities = frame.GetEntities()

		case MVDSvcConfigString:
			cs, err := p.ParseConfigString(msg)
			if err != nil {
				return nil, err
			}
			p.demo.Configstrings[int32(cs.GetIndex())] = cs
			if cbFunc, found := p.callbacks[MVDSvcConfigString]; found {
				cbFunc(cs)
			}

			// is it a player skin? then add the player
			playerNum := int32(cs.GetIndex()) - p.demo.Remap.PlayerSkins
			if 0 <= playerNum && playerNum <= p.demo.MaxPlayers {
				p.demo.Players[playerNum] = &pb.MvdPlayer{
					Name: cs.Data[:strings.Index(cs.Data, "\\")],
				}
			}

		case MVDSvcFrame:
			frame, err := p.ParseFrame(msg)
			if err != nil {
				return nil, err
			}
			if cbFunc, found := p.callbacks[MVDSvcFrame]; found {
				cbFunc(frame)
			}
			packet.Frames = append(packet.Frames, frame)
			p.demo.Entities = frame.GetEntities() // maybe??

		case MVDSvcSound:
			sound := p.ParseSound(msg, extra)
			packet.Sounds = append(packet.Sounds, sound)
			if cbFunc, found := p.callbacks[MVDSvcSound]; found {
				cbFunc(sound)
			}

		case MVDSvcPrint:
			print := &pb.Print{
				Level: uint32(msg.ReadByte()),
				Data:  msg.ReadString(),
			}
			packet.Prints = append(packet.Prints, print)
			if cbFunc, found := p.callbacks[MVDSvcPrint]; found {
				cbFunc(print)
			}

		case MVDSvcUnicast:
			fallthrough
		case MVDSvcUnicastReliable:
			reliable := cmd == MVDSvcUnicastReliable
			unicast, err := p.ParseUnicast(msg, reliable, extra)
			if err != nil {
				return nil, err
			}
			packet.Unicasts = append(packet.Unicasts, unicast)
			if cbFunc, found := p.callbacks[MVDSvcUnicast]; found {
				cbFunc(unicast)
			}
			if cbFunc, found := p.callbacks[MVDSvcUnicastReliable]; found {
				cbFunc(unicast)
			}

		case MVDSvcMulticastAll:
			fallthrough
		case MVDSvcMulticastAllR:
			fallthrough
		case MVDSvcMulticastPHS:
			fallthrough
		case MVDSvcMulticastPHSR:
			fallthrough
		case MVDSvcMulticastPVS:
			fallthrough
		case MVDSvcMulticastPVSR:
			multicast := p.ParseMulticast(msg, int(cmd)-MVDSvcMulticastAll, extra)
			packet.Multicasts = append(packet.Multicasts, multicast)
		}
	}
	return packet, nil
}

// ServerData is the first message in a demo, it contains info about what
// protocol versions were used at the time of the capture, the game directory,
// the client number of the dummy spec, and more.
func (p *MVD2Parser) ParseServerData(msg *message.Buffer, extra int) error {
	p.index++
	p.allDemos = append(p.allDemos, &pb.MvdDemo{})
	demo := p.allDemos[p.index]
	p.demo = p.allDemos[p.index] // set the pointer to the new current demo

	if msg.ReadLongP() != 37 {
		return fmt.Errorf("parse error: demo protocol not 37")
	}
	demo.Version = msg.ReadShortP()

	if demo.Version >= ProtocolPlusPlus {
		demo.Flags = msg.ReadShortP()
	} else {
		demo.Flags = int32(extra)
	}

	demo.Identity = msg.ReadLongP()
	demo.GameDir = msg.ReadString()
	demo.Dummy = msg.ReadShortP()

	demo.EntityStateFlags = EntityStateUMask | EntityStateBeamOrigin
	demo.Remap = csRemap

	if (demo.Version >= ProtocolPlus) && ((demo.Flags & FlagExtLimits) != 0) {
		demo.EntityStateFlags |= EntityStateLongSolid | EntityStateShortAngles | EntityStateExtensions
		demo.PlayerStateFlags |= PlayerStateExtensions
		demo.Remap = csRemapNew
	}

	if (demo.Version >= ProtocolPlusPlus) && ((demo.Flags & FlagExtLimits2) != 0) {
		demo.EntityStateFlags |= EntityStateExtensions2
		demo.PlayerStateFlags |= PlayerStateExtensions2
		if demo.Version >= ProtocolPlayerFog {
			demo.PlayerStateFlags |= MvdPlayerMoreBits
		}
	}

	if p.debug {
		fmt.Printf("serverdata - ver:%d, dummy:%d, ident:%d game:%q\n", demo.GetVersion(), demo.GetDummy(), demo.GetIdentity(), demo.GetGameDir())
	}
	return nil
}

// This parses the blob of configstrings that directly follow the serverdata at
// the beginning of the demo file.
//
// The list of players in the demo is built using configstrings for skins. The
// cs_index - skin offset == the player number (0-maxclients), and the player
// name is the prefix of the skin string.
func (p *MVD2Parser) ParseConfigStrings(msg *message.Buffer) map[int32]*pb.ConfigString {
	out := make(map[int32]*pb.ConfigString)
	for {
		if msg.Index >= msg.Length {
			break
		}
		idx := int32(msg.ReadShort())
		if idx == int32(p.demo.Remap.GetEnd()) {
			break
		}
		out[idx] = &pb.ConfigString{Index: uint32(idx), Data: msg.ReadString()}
		if p.debug {
			fmt.Printf("configstring - [%d] %q\n", idx, out[idx].GetData())
		}
	}
	mc, ok := out[p.demo.Remap.GetMaxClients()]
	if ok {
		maxclients, _ := strconv.Atoi(mc.Data)
		p.demo.MaxPlayers = int32(maxclients)
		p.demo.Players = make(map[int32]*pb.MvdPlayer)
		for i := int32(0); i < p.demo.MaxPlayers; i++ {
			cs, ok := out[p.demo.Remap.GetPlayerSkins()+i]
			if ok {
				p.demo.Players[i] = &pb.MvdPlayer{
					Name: cs.Data[:strings.Index(cs.Data, "\\")],
				}
			}
		}
	}
	return out
}

// Parse a single configstring from MVD data
func (p *MVD2Parser) ParseConfigString(msg *message.Buffer) (*pb.ConfigString, error) {
	if msg.Index >= msg.Length {
		return nil, fmt.Errorf("ParseConfigString() error - end of buffer")
	}
	idx := msg.ReadShort()
	if idx >= int(int16(p.demo.Remap.End)) {
		return nil, fmt.Errorf("ParseConfigString() error - index out of bounds: %d", idx)
	}
	str := msg.ReadString()
	if p.debug {
		fmt.Printf("configstring - [%d] %q\n", idx, str)
	}
	return &pb.ConfigString{
		Index: uint32(idx),
		Data:  str,
	}, nil
}

// Each frame contains portal data, then all the player POVs (delta compressed),
// and then all the compressed entities.
func (p *MVD2Parser) ParseFrame(msg *message.Buffer) (*pb.MvdFrame, error) {
	frame := &pb.MvdFrame{}
	frame.PortalData = msg.ReadData(msg.ReadByte())

	players, err := p.ParsePacketPlayers(msg)
	if err != nil {
		return nil, err
	}
	frame.Players = players
	ents, err := p.ParseDeltaEntities(msg)
	if err != nil {
		return nil, err
	}
	frame.Entities = ents
	if p.debug {
		fmt.Printf("frame - portal data:%v\n", frame.PortalData)
	}
	return frame, nil
}

// Read all the player info from a frame.
func (p *MVD2Parser) ParsePacketPlayers(msg *message.Buffer) (map[int32]*pb.PackedPlayer, error) {
	var bits uint32
	out := make(map[int32]*pb.PackedPlayer)
	for {
		number := int32(msg.ReadByte())
		if number == ClientNumNone {
			break
		}
		pl, ok := p.demo.Players[number]
		if !ok {
			//return nil, fmt.Errorf("ParsePacketPlayers(%d) - player not found", number)
			pl = &pb.MvdPlayer{
				Name: "unknown",
			}
		}
		// check num bounds later
		bits = uint32(msg.ReadWord())
		ps, err := p.ParseDeltaPlayer(msg, bits, p.demo.PlayerStateFlags)
		if err != nil {
			return nil, fmt.Errorf("error parsing player: %v", err)
		}
		pl.PlayerState = ps

		if (bits & MvdPlayerRemove) != 0 {
			pl.InUse = false
			continue
		}
		pl.InUse = true
		out[number] = ps
	}
	if p.debug {
		var names []string
		for k := range out {
			names = append(names, p.demo.GetPlayers()[k].GetName())
		}
		fmt.Printf("players - %q\n", strings.Join(names, ", "))
	}
	return out, nil
}

// Parse a compressed player. Parsing delta players from regular DM2 demos is
// similar but not identical, so a separate func is needed.
func (p *MVD2Parser) ParseDeltaPlayer(msg *message.Buffer, bits uint32, flags int32) (*pb.PackedPlayer, error) {
	to := &pb.PackedPlayer{}
	pm := &pb.PlayerMove{}
	if (bits & MvdPlayerType) != 0 {
		pm.Type = msg.ReadByteP()
	}
	if flags&MvdPlayerFlagExtensions2 != 0 {
		log.Println("MVD Playerstate Extensions found")
	} else {
		if (bits & MvdPlayerOrigin) != 0 {
			pm.OriginX = msg.ReadShortP()
			pm.OriginY = msg.ReadShortP()
		}
		if (bits & MvdPlayerOrigin2) != 0 {
			pm.OriginZ = msg.ReadShortP()
		}
	}
	if (bits & MvdPlayerViewOffset) != 0 {
		to.ViewOffsetX = msg.ReadCharP()
		to.ViewOffsetY = msg.ReadCharP()
		to.ViewOffsetZ = msg.ReadCharP()
	}
	if (bits & MvdPlayerViewAngles) != 0 {
		to.ViewAnglesX = msg.ReadShortP()
		to.ViewAnglesY = msg.ReadShortP()
	}
	if (bits & MvdPlayerViewAngles2) != 0 {
		to.ViewAnglesZ = msg.ReadShortP()
	}
	if (bits & MvdPlayerKickAngles) != 0 {
		to.KickAnglesX = msg.ReadCharP()
		to.KickAnglesY = msg.ReadCharP()
		to.KickAnglesZ = msg.ReadCharP()
	}
	if (bits & MvdPlayerWeaponIndex) != 0 {
		if (flags & MvdPlayerFlagExtensions) != 0 {
			to.GunIndex = msg.ReadWordP()
		} else {
			to.GunIndex = msg.ReadByteP()
		}
	}
	if (bits & MvdPlayerWeaponFrame) != 0 {
		to.GunFrame = msg.ReadByteP()
	}
	if (bits & MvdPlayerGunOffset) != 0 {
		to.GunOffsetX = msg.ReadCharP()
		to.GunOffsetY = msg.ReadCharP()
		to.GunOffsetZ = msg.ReadCharP()
	}
	if (bits & MvdPlayerGunAngles) != 0 {
		to.GunAnglesX = msg.ReadCharP()
		to.GunAnglesY = msg.ReadCharP()
		to.GunAnglesZ = msg.ReadCharP()
	}
	if (bits & MvdPlayerBlend) != 0 {
		if (flags & MvdPlayerFlagExtensions2) != 0 {
			bf := msg.ReadByte()
			if (bf & (1 << 0)) != 0 {
				to.BlendW = int32(msg.ReadByte())
			}
			if (bf & (1 << 1)) != 0 {
				to.BlendX = int32(msg.ReadByte())
			}
			if (bf & (1 << 2)) != 0 {
				to.BlendY = int32(msg.ReadByte())
			}
			if (bf & (1 << 3)) != 0 {
				to.BlendZ = int32(msg.ReadByte())
			}
			if (bf & (1 << 0)) != 0 {
				to.DamageBlendW = int32(msg.ReadByte())
			}
			if (bf & (1 << 1)) != 0 {
				to.DamageBlendX = int32(msg.ReadByte())
			}
			if (bf & (1 << 2)) != 0 {
				to.DamageBlendY = int32(msg.ReadByte())
			}
			if (bf & (1 << 3)) != 0 {
				to.DamageBlendZ = int32(msg.ReadByte())
			}
		} else {
			to.BlendW = int32(msg.ReadByte())
			to.BlendX = int32(msg.ReadByte())
			to.BlendY = int32(msg.ReadByte())
			to.BlendZ = int32(msg.ReadByte())
		}
	}
	if (bits & MvdPlayerFov) != 0 {
		to.Fov = msg.ReadByteP()
	}
	if (bits & MvdPlayerRdFlags) != 0 {
		to.RdFlags = msg.ReadByteP()
	}
	if (bits & MvdPlayerStats) != 0 {
		stats := p.ParsePlayerStats(msg, flags)
		to.Stats = stats
	}
	to.Movestate = pm
	return to, nil
}

// Stats are a set of integer values at the end of each playerstate. They are
// are used to transfer rapidly changing values from server to client for
// displaying in the HUD. Score, health value, armor value, etc. These are
// transfered every server frame.
//
// Some stats numbers are direct values (health, armor), and some are indexes
// to things like config string values.
func (p *MVD2Parser) ParsePlayerStats(msg *message.Buffer, flags int32) map[uint32]int32 {
	stats := make(map[uint32]int32)
	var bits uint64
	var num uint32
	if (flags & MvdPlayerFlagExtensions2) != 0 {
		bits = msg.ReadVarInt64()
		num = MaxStatsExt
	} else {
		bits = uint64(msg.ReadLong())
		num = MaxStats
	}
	if bits == 0 {
		return nil
	}
	for i := uint32(0); i < num; i++ {
		if (bits & (1 << i)) != 0 {
			stats[i] = msg.ReadShortP()
		}
	}
	return stats
}

// Parse all the entities from a frame. These come directly after all the playerstates.
func (p *MVD2Parser) ParseDeltaEntities(msg *message.Buffer) (map[int32]*pb.PackedEntity, error) {
	var bits int64
	var num int32
	var err error

	if p.demo.Entities == nil {
		p.demo.Entities = make(map[int32]*pb.PackedEntity)
	}

	for {
		num, bits = p.ParseEntityBits(msg)
		if num == 0 {
			break
		}
		if bits == 0 {
			continue
		}
		ent := p.demo.Entities[num]
		ent, err = p.ParseDeltaEntity(msg, bits, ent)
		if err != nil {
			return nil, err
		}
		if (bits & message.EntityRemove) != 0 {
			if (ent.RenderFx & message.RFBeam) == 0 {
				ent.OldOriginX = ent.GetOriginX()
				ent.OldOriginY = ent.GetOriginY()
				ent.OldOriginZ = ent.GetOriginZ()
			}
			// set inuse false
		}
		ent.Number = uint32(num)
		p.demo.Entities[num] = ent
	}
	if p.debug {
		fmt.Printf("entities\n")
	}
	return nil, nil
}

// Each entity is prefixed with up to 5 bytes of bitmask followed by the entity
// number (up to 2 bytes) and then the actual data.
func (p *MVD2Parser) ParseEntityBits(msg *message.Buffer) (int32, int64) {
	number := int32(0)
	if msg.Index == msg.Length {
		return 0, 0
	}

	mask := int64(msg.ReadByte())
	if (mask & message.EntityMoreBits1) != 0 {
		mask |= (int64(msg.ReadByte()) << 8)
	}
	if (mask & message.EntityMoreBits2) != 0 {
		mask |= (int64(msg.ReadByte()) << 16)
	}
	if (mask & message.EntityMoreBits3) != 0 {
		mask |= (int64(msg.ReadByte()) << 24)
	}
	if (p.demo.EntityStateFlags & EntityStateExtensions) != 0 {
		if (mask & message.EntityMoreBits4) != 0 {
			mask |= (int64(msg.ReadByte()) << 32)
		}
	}

	if (mask & message.EntityNumber16) != 0 {
		number = int32(msg.ReadWord())
	} else {
		number = int32(msg.ReadByte())
	}
	return number, mask
}

// Parse an individual edict_t. These are delta compressed (only changes from
// the last update are emitted); the PackedEntity proto returned merged with
// the previos version of this entity.
func (p *MVD2Parser) ParseDeltaEntity(msg *message.Buffer, bits int64, from *pb.PackedEntity) (*pb.PackedEntity, error) {
	flags := p.demo.EntityStateFlags
	to := &pb.PackedEntity{}
	if bits == 0 {
		return nil, fmt.Errorf("error parsing entity, bitmask == 0")
	}

	if from != nil {
		to = proto.Clone(from).(*pb.PackedEntity)
	}

	if ((flags & EntityStateExtensions) != 0) && ((bits & message.EntityModel16) != 0) {
		if (bits & message.EntityModel) != 0 {
			to.ModelIndex = uint32(msg.ReadWord())
		}
		if (bits & message.EntityModel2) != 0 {
			to.ModelIndex2 = uint32(msg.ReadWord())
		}
		if (bits & message.EntityModel3) != 0 {
			to.ModelIndex3 = uint32(msg.ReadWord())
		}
		if (bits & message.EntityModel4) != 0 {
			to.ModelIndex4 = uint32(msg.ReadWord())
		}
	} else {
		if (bits & message.EntityModel) != 0 {
			to.ModelIndex = uint32(msg.ReadByte())
		}
		if (bits & message.EntityModel2) != 0 {
			to.ModelIndex2 = uint32(msg.ReadByte())
		}
		if (bits & message.EntityModel3) != 0 {
			to.ModelIndex3 = uint32(msg.ReadByte())
		}
		if (bits & message.EntityModel4) != 0 {
			to.ModelIndex4 = uint32(msg.ReadByte())
		}
	}

	if (bits & message.EntityFrame8) != 0 {
		to.Frame = uint32(msg.ReadByte())
	}

	if (bits & message.EntityFrame16) != 0 {
		to.Frame = uint32(msg.ReadWord())
	}

	if (bits & message.EntitySkin32) == message.EntitySkin32 {
		to.Skin = uint32(msg.ReadLong())
	} else if (bits & message.EntitySkin8) != 0 {
		to.Skin = uint32(msg.ReadByte())
	} else if (bits & message.EntitySkin16) != 0 {
		to.Skin = uint32(msg.ReadWord())
	}

	if (bits & message.EntityEffects32) == message.EntityEffects32 {
		to.Effects = uint32(msg.ReadLong())
	} else if (bits & message.EntityEffects8) != 0 {
		to.Effects = uint32(msg.ReadByte())
	} else if (bits & message.EntityEffects16) != 0 {
		to.Effects = uint32(msg.ReadWord())
	}

	if (bits & message.EntityRenderFX32) == message.EntityRenderFX32 {
		to.RenderFx = uint32(msg.ReadLong())
	} else if (bits & message.EntityRenderFX8) != 0 {
		to.RenderFx = uint32(msg.ReadByte())
	} else if (bits & message.EntityRenderFX16) != 0 {
		to.RenderFx = uint32(msg.ReadWord())
	}

	if (flags & EntityStateExtensions2) != 0 {
		// read delta coords for origins here
	} else {
		if (bits & message.EntityOrigin1) != 0 {
			to.OriginX = int32(msg.ReadShort())
		}
		if (bits & message.EntityOrigin2) != 0 {
			to.OriginY = int32(msg.ReadShort())
		}
		if (bits & message.EntityOrigin3) != 0 {
			to.OriginZ = int32(msg.ReadShort())
		}
	}

	if ((flags & EntityStateShortAngles) != 1) && ((bits & message.EntityAngle16) != 0) {
		if (bits & message.EntityAngle1) != 0 {
			to.AngleX = int32(msg.ReadShort())
		}
		if (bits & message.EntityAngle2) != 0 {
			to.AngleY = int32(msg.ReadShort())
		}
		if (bits & message.EntityAngle3) != 0 {
			to.AngleZ = int32(msg.ReadShort())
		}
	} else {
		if (bits & message.EntityAngle1) != 0 {
			to.AngleX = int32(msg.ReadChar())
		}
		if (bits & message.EntityAngle2) != 0 {
			to.AngleY = int32(msg.ReadChar())
		}
		if (bits & message.EntityAngle3) != 0 {
			to.AngleZ = int32(msg.ReadChar())
		}
	}

	if (bits & message.EntityOldOrigin) != 0 { // if extended2 read delta coords
		to.OldOriginX = int32(msg.ReadShort())
		to.OldOriginY = int32(msg.ReadShort())
		to.OldOriginZ = int32(msg.ReadShort())
	}

	if (bits & message.EntitySound) != 0 { // if extensions do more
		to.Sound = uint32(msg.ReadByte())
	}

	if (bits & message.EntityEvent) != 0 {
		to.Event = uint32(msg.ReadByte())
	}

	if (bits & message.EntitySolid) != 0 {
		if (flags & EntityStateLongSolid) != 0 {
			to.Solid = uint32(msg.ReadLong())
		} else {
			to.Solid = uint32(msg.ReadWord())
		}
	}

	if (flags & EntityStateExtensions) != 0 {
		if p.demo.Extension == nil {
			p.demo.Extension = &pb.MvdEntityStateExtension{}
		}
		if (bits & message.EntityMoreFX32) == message.EntityMoreFX32 {
			p.demo.Extension.MoreFx = int32(msg.ReadLong())
		} else if (bits & message.EntityMoreFX8) != 0 {
			p.demo.Extension.MoreFx = int32(msg.ReadByte())
		} else if (bits & message.EntityMoreFX16) != 0 {
			p.demo.Extension.MoreFx = int32(msg.ReadWord())
		}

		if (bits & message.EntityAlpha) != 0 {
			p.demo.Extension.Alpha = int32(msg.ReadByte())
		}

		if (bits & message.EntityScale) != 0 {
			p.demo.Extension.Scale = int32(msg.ReadByte())
		}
	}

	return to, nil
}

func (p *MVD2Parser) ParseUnicast(msg *message.Buffer, reliable bool, extra int) (*pb.MvdUnicast, error) {
	out := &pb.MvdUnicast{}
	len := msg.ReadByteP()
	len |= uint32(extra) << 8

	clientNum := int32(msg.ReadByteP())
	if clientNum >= p.demo.Remap.GetMaxClients() {
		return nil, fmt.Errorf("ParseUnicast error - client more than max: %d", clientNum)
	}
	player, ok := p.demo.GetPlayers()[clientNum]
	if !ok {
		player = &pb.MvdPlayer{Name: "[unknown]", InUse: true}
	}
	out.ClientNumber = clientNum
	out.Player = player

	readStart := msg.Index
	for {
		if (msg.Index - readStart) >= int(len) {
			break
		}
		cmd := msg.ReadByte()
		switch cmd {
		case SvcLayout:
			layout := &pb.Layout{Data: msg.ReadString()}
			out.Layouts = append(out.Layouts, layout)
		case SvcConfigString:
			cs := &pb.ConfigString{Index: uint32(msg.ReadWord()), Data: msg.ReadString()}
			out.Configstrings = append(out.Configstrings, cs)
		case SvcPrint:
			p := &pb.Print{Level: uint32(msg.ReadByte()), Data: msg.ReadString()}
			out.Prints = append(out.Prints, p)
		case SvcStuffText:
			st := &pb.StuffText{Data: msg.ReadString()}
			out.Stuffs = append(out.Stuffs, st)
		}
	}
	if p.debug {
		fmt.Printf("unicast - [%d] %q\n", clientNum, player.Name)
		for _, s := range out.GetStuffs() {
			fmt.Printf("  stuff - %q\n", s.Data)
		}
		for _, c := range out.GetConfigstrings() {
			fmt.Printf("  configstring - [%d] %q\n", c.GetIndex(), c.GetData())
		}
		for _, lo := range out.GetLayouts() {
			fmt.Printf("  layout - %q\n", lo.GetData())
		}
		for _, p := range out.GetPrints() {
			fmt.Printf("  print - [%d] %q\n", p.GetLevel(), p.GetData())
		}
	}
	return out, nil
}

func (p *MVD2Parser) ParseSound(msg *message.Buffer, extra int) *pb.PackedSound {
	var index uint32
	s := &pb.PackedSound{}
	flags := msg.ReadByteP()
	s.Flags = flags
	if p.demo.Remap.Extended && ((flags & message.SoundIndex16) != 0) {
		index = msg.ReadWordP()
	} else {
		index = msg.ReadByteP()
	}
	s.Index = index

	if (flags & message.SoundVolume) != 0 {
		s.Volume = msg.ReadByteP()
	}
	if (flags & message.SoundAttenuation) != 0 {
		s.Attenuation = msg.ReadByteP()
	}
	if (flags & message.SoundOffset) != 0 {
		s.TimeOffset = msg.ReadByteP()
	}

	sendchan := msg.ReadWordP()
	entnum := int32(sendchan >> 3)
	ent := p.demo.Entities[entnum]
	s.Entity = ent.GetNumber()
	if p.debug {
		fmt.Printf(
			"sound - [%d] %q\n",
			s.GetIndex(),
			p.demo.GetConfigstrings()[p.demo.GetRemap().GetSounds()+int32(s.GetIndex())].Data,
		)
	}
	return s
}

// ParseMulticast is used to parse all 6 multicast cmd types
func (p *MVD2Parser) ParseMulticast(msg *message.Buffer, to int, extra int) *pb.MvdMulticast {
	out := &pb.MvdMulticast{}
	len := msg.ReadByteP()
	len |= uint32(extra) << 8
	if to != 0 {
		out.Leaf = int32(msg.ReadWordP())
	}
	out.Data = msg.ReadData(int(len))

	if cbFunc, found := p.callbacks[to]; found {
		cbFunc(out)
	}
	if p.debug {
		fmt.Printf("multicast - %v\n", out.GetData())
	}
	return out
}

// RegisterCallback allows for a custom function to be called at specific
// points while a demo is being parsed.
//
// `event` is an index corresponding to what you want your callback to be
// associated with. These match up with the SVC* server messages.
//
// `dofunc` is the function definition to be called at that event. The argument
// is set to `any` to accept any type, but this arg should be the parsed server
// message proto. Inside the dofunc, the arg will need to be casted to the
// appropriate type. Something like:
//
// printMsg := argname.(*pb.Print)
func (p *MVD2Parser) RegisterCallback(event int, dofunc func(any)) {
	if p.callbacks == nil {
		p.callbacks = make(map[int]func(any))
	}
	p.callbacks[event] = dofunc
}

// Dynamically remove a particular callback
func (p *MVD2Parser) UnregisterCallback(msgtype int) {
	delete(p.callbacks, msgtype)
}

// Private member accessor
func (p *MVD2Parser) GetRawData() []byte {
	return p.binaryData
}

// Private member accessor
func (p *MVD2Parser) GetRawPosition() int {
	return p.binaryPosition
}

// Private member accessor
func (p *MVD2Parser) IsZipped() bool {
	return p.zipped
}

// Private member accessor
func (p *MVD2Parser) GetDemos() []*pb.MvdDemo {
	return p.allDemos
}

// Private member accessor
func (p *MVD2Parser) GetDebug() bool {
	return p.debug
}
