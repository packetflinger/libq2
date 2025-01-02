package message

import (
	"errors"
	"strconv"
	"strings"

	pb "github.com/packetflinger/libq2/proto"
)

const (
	MaxItems         = 256
	MaxStats         = 32
	MaxEntities      = 1024
	MaxConfigStrings = 2080
	MaxMessageLength = 1390
	CLCMove          = 2
	CLCUserinfo      = 3
	CLCStringCommand = 4
	PrintLevelLow    = 1
	PrintLevelObit   = 2
	PrintLevelHigh   = 3
	PrintLevelChat   = 3
)

const (
	CallbackOnConnect = iota + SVCNumTypes
	CallbackOnBegin
	CallbackOnMention // a print containing our name
	CallbackOnDamage  // someone shot us
	CallbackOnDie     // we're dead
)

type Parser interface {
	ApplyPacket(packet *pb.Packet)
}

// The ServerData message is the first one sent from the server to the client
// after the client connects. It contains info about the protocol used, the
// current map name (not the map filename), etc.
func (m *Buffer) ParseServerData() *pb.ServerInfo {
	if m.Index == m.Length {
		return nil
	}
	return &pb.ServerInfo{
		Protocol:     m.ReadULong(),
		ServerCount:  m.ReadULong(),
		Demo:         m.ReadByte() == 1,
		GameDir:      m.ReadString(),
		ClientNumber: uint32(m.ReadShort()),
		MapName:      m.ReadString(),
	}
}

// Configstrings are strings sent to each client and associated
// with an index. They're referenced by index in various playces
// such as layouts, etc.
func (m *Buffer) ParseConfigString() *pb.ConfigString {
	if m.Index == m.Length {
		return nil
	}
	return &pb.ConfigString{
		Index: uint32(m.ReadShort()),
		Data:  m.ReadString(),
	}
}

// A baseline is just a normal entity in its default state, from
// a client's perspective
func (m *Buffer) ParseSpawnBaseline() *pb.PackedEntity {
	bitmask := m.ParseEntityBitmask()
	number := m.ParseEntityNumber(bitmask)
	return m.ParseEntity(&pb.PackedEntity{}, number, bitmask)
}

// Stuffs are a way for the server to force a client to execute a command. If
// a client is connected to a server, it will do anything the server tells it
// to via a StuffText message. Anything a player can type into their console
// can be stuff'd to them, including harmful things like `disconnect` or
// `set rate "10"`.
func (m *Buffer) ParseStuffText() *pb.StuffText {
	if m.Index == m.Length {
		return nil
	}
	return &pb.StuffText{Data: m.ReadString()}
}

// Frame message are sent from the server to each client every 0.1 seconds (100
// milliseconds), this is the server's default framerate. Each fram will
// contain info about that frame (the number, the number of the frame it was
// compressed against, areabits, etc), and also the current playerstate and
// copy of all entities that changed since the delta frame.
func (m *Buffer) ParseFrame(oldFrames map[int32]*pb.Frame) *pb.Frame {
	if m.Index == m.Length {
		return nil
	}
	var fromPS *pb.PackedPlayer
	var fromEnts map[int32]*pb.PackedEntity
	fr := &pb.Frame{}
	fr.Number = int32(m.ReadLong())
	fr.Delta = int32(m.ReadLong())
	fr.Suppressed = uint32(m.ReadByte())
	fr.AreaBytes = uint32(m.ReadByte())
	areabits := m.ReadData(int(fr.GetAreaBytes()))
	for _, ab := range areabits {
		fr.AreaBits = append(fr.AreaBits, uint32(ab))
	}
	if oldFrames != nil {
		delta, ok := oldFrames[fr.Delta]
		if ok {
			fromPS = delta.GetPlayerState()
			fromEnts = delta.GetEntities()
		}
	}
	var ps *pb.PackedPlayer
	if m.ReadByte() == SVCPlayerInfo {
		ps = m.ParseDeltaPlayerstate(fromPS)
	}
	fr.PlayerState = ps
	if m.ReadByte() == SVCPacketEntities {
		fr.Entities = m.ParsePacketEntities(fromEnts)
	}
	return fr
}

// Print messages are sent from the server to all clients when there is
// something that should be printed in the console of all clients. This
// includes player chat, obituaries, inventory (out of ammo, cant switch to
// the railgun).
func (m *Buffer) ParsePrint() *pb.Print {
	if m.Index == m.Length {
		return nil
	}
	return &pb.Print{
		Level: uint32(m.ReadByte()),
		Data:  m.ReadString(),
	}
}

// This is a start-sound packet
func (m *Buffer) ParseSound() *pb.PackedSound {
	s := &pb.PackedSound{}
	s.Flags = uint32(m.ReadByte())
	s.Index = uint32(m.ReadByte())
	if (s.Flags & SoundVolume) > 0 {
		s.Volume = uint32(m.ReadByte())
	} else {
		s.Volume = 1
	}
	if (s.Flags & SoundAttenuation) > 0 {
		s.Attenuation = uint32(m.ReadByte())
	} else {
		s.Attenuation = 1
	}
	if (s.Flags & SoundOffset) > 0 {
		s.TimeOffset = uint32(m.ReadByte())
	} else {
		s.TimeOffset = 0
	}
	if (s.Flags & SoundEntity) > 0 {
		tmp := uint32(m.ReadShort())
		s.Entity = tmp >> 3
		s.Channel = tmp & 7
	} else {
		s.Channel = 0
		s.Entity = 0
	}
	if (s.Flags & SoundPosition) > 0 {
		pos := m.ReadPosition()
		s.PositionX, s.PositionY, s.PositionZ = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	}
	return s
}

// Temporary entities include things like explosions, sparks, and things like
// projectiles (rockets).
func (m *Buffer) ParseTempEntity() *pb.TemporaryEntity {
	if m.Index == m.Length {
		return nil
	}
	te := &pb.TemporaryEntity{}
	te.Type = uint32(m.ReadByte())
	switch te.Type {
	case TentBlood:
		fallthrough
	case TentGunshot:
		fallthrough
	case TentSparks:
		fallthrough
	case TentBulletSparks:
		fallthrough
	case TentScreenSparks:
		fallthrough
	case TentShieldSparks:
		fallthrough
	case TentShotgun:
		fallthrough
	case TentBlaster:
		fallthrough
	case TentGreenBlood:
		fallthrough
	case TentBlaster2:
		fallthrough
	case TentFlechette:
		fallthrough
	case TentHeatBeamSparks:
		fallthrough
	case TentHeatBeamSteam:
		fallthrough
	case TentMoreBlood:
		fallthrough
	case TentElectricSparks:
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		te.Direction = uint32(m.ReadDirection())
	case TentSplash:
		fallthrough
	case TentLaserSparks:
		fallthrough
	case TentWeldingSparks:
		fallthrough
	case TentTunnelSparks:
		te.Count = uint32(m.ReadByte())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		te.Direction = uint32(m.ReadDirection())
		te.Color = uint32(m.ReadByte())
	case TentBlueHyperBlaster:
		fallthrough
	case TentRailTrail:
		fallthrough
	case TentBubbleTrail:
		fallthrough
	case TentDebugTrail:
		fallthrough
	case TentBubbleTrail2:
		fallthrough
	case TentBFGLaser:
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.Position2X, te.Position2Y, te.Position2Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	case TentGrenadeExplosion:
		fallthrough
	case TentGrenadeExplosionWater:
		fallthrough
	case TentExplosion2:
		fallthrough
	case TentPlasmaExplosion:
		fallthrough
	case TentRocketExplosion:
		fallthrough
	case TentRocketExplosionWater:
		fallthrough
	case TentExplosion1:
		fallthrough
	case TentExplosion1NP:
		fallthrough
	case TentExplosion1Big:
		fallthrough
	case TentBFGExplosion:
		fallthrough
	case TentBFGBigExplosion:
		fallthrough
	case TentBossTeleport:
		fallthrough
	case TentPlainExplosion:
		fallthrough
	case TentChainFistSmoke:
		fallthrough
	case TentTrackerExplosion:
		fallthrough
	case TentTeleportEffect:
		fallthrough
	case TentDBallGoal:
		fallthrough
	case TentWidowSplash:
		fallthrough
	case TentNukeBlast:
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	case TentParasiteAttack:
		fallthrough
	case TentMedicCableAttack:
		fallthrough
	case TentHeatBeam:
		fallthrough
	case TentMonsterHeatBeam:
		te.Entity1 = int32(m.ReadShort())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.Position2X, te.Position2Y, te.Position2Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.OffsetX, te.OffsetY, te.OffsetZ = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	case TentGrappleCable:
		te.Entity1 = int32(m.ReadShort())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.Position2X, te.Position2Y, te.Position2Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.OffsetX, te.OffsetY, te.OffsetZ = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	case TentLightning:
		te.Entity1 = int32(m.ReadShort())
		te.Entity2 = int32(m.ReadShort())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.Position2X, te.Position2Y, te.Position2Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	case TentFlashlight:
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		te.Entity1 = int32(m.ReadShort())
	case TentForceWall:
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		pos = m.ReadPosition()
		te.Position2X, te.Position2Y, te.Position2Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		te.Color = uint32(m.ReadByte())
	case TentSteam:
		te.Entity1 = int32(m.ReadShort())
		te.Count = uint32(m.ReadByte())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
		te.Direction = uint32(m.ReadDirection())
		te.Color = uint32(m.ReadByte())
		te.Entity2 = int32(m.ReadShort())
		if te.Entity1 != -1 {
			te.Time = int32(m.ReadLong())
		}
	case TentWidowBeamOut:
		te.Entity1 = int32(m.ReadShort())
		pos := m.ReadPosition()
		te.Position1X, te.Position1Y, te.Position1Z = uint32(pos[0]), uint32(pos[1]), uint32(pos[2])
	}

	return te
}

// A gun fired, nearby clients should see the flash. Monsters also produce
// muzzle flashes when they shoot, but only in single-player mode.
func (m *Buffer) ParseMuzzleFlash() *pb.MuzzleFlash {
	if m.Index == m.Length {
		return nil
	}
	return &pb.MuzzleFlash{
		Entity: uint32(m.ReadShort()),
		Weapon: uint32(m.ReadByte()),
	}
}

// A layout is a string of code to represent how things need
// to be arranged on the screen. The intermission screen
// after a TDM match for example with players, scores, pings,
// stats, etc is an example
func (m *Buffer) ParseLayout() *pb.Layout {
	if m.Index == m.Length {
		return nil
	}
	return &pb.Layout{
		Data: m.ReadString(),
	}
}

// 2 bytes for every item
func (m *Buffer) ParseInventory() {
	// we don't actually care about this, advance the buffer's pointer so we
	// can accurately find any messages after this one.
	m.Index += 2 * MaxItems
}

// A string that should appear temporarily in the center of the screen
func (m *Buffer) ParseCenterPrint() *pb.CenterPrint {
	if m.Index == m.Length {
		return nil
	}
	return &pb.CenterPrint{
		Data: m.ReadString(),
	}
}

// This the first server-to-client message after client issues "getchallenge"
// when initiating a connection.
//
// Example response: ÿÿÿÿchallenge 910908644 p=34,35,36
func (m *Buffer) ParseChallenge() (*pb.Challenge, error) {
	cl := m.ReadLong()
	if cl != -1 {
		return nil, errors.New("not connectionless message, invalid challenge response")
	}
	line := m.ReadString()
	tokens := strings.Fields(line)
	chal := int32(0)
	if len(tokens) > 1 {
		num, err := strconv.Atoi(tokens[1])
		if err != nil {
			return nil, errors.New("invalid challenge response")
		}
		chal = int32(num)
	}
	pr := []int32{}
	if len(tokens) > 2 {
		protocols := strings.Split(tokens[2][2:], ",")
		for _, p := range protocols {
			pint, err := strconv.Atoi(p)
			if err != nil {
				continue
			}
			pr = append(pr, int32(pint))
		}
	}
	return &pb.Challenge{
		Number:    chal,
		Protocols: pr,
	}, nil
}

/*
func (msg *MessageBuffer) WriteDeltaMove(from *client.ClientMove, to *client.ClientMove) {
	bits := byte(0)
	if to.Angles.X != from.Angles.X {
		bits |= client.MoveAngle1
	}
	if to.Angles.Y != from.Angles.Y {
		bits |= client.MoveAngle2
	}
	if to.Angles.Z != from.Angles.Z {
		bits |= client.MoveAngle3
	}
	if to.Forward != from.Forward {
		bits |= client.MoveForward
	}
	if to.Sideways != from.Sideways {
		bits |= client.MoveSide
	}
	if to.Up != from.Up {
		bits |= client.MoveUp
	}
	if to.Buttons != from.Buttons {
		bits |= client.MoveButtons
	}
	if to.Impulse != from.Impulse {
		bits |= client.MoveImpulse
	}

	msg.WriteByte(bits)
	buttons := from.Buttons & client.ButtonMask

	if bits&client.MoveAngle1 > 0 {
		msg.WriteShort(uint16(to.Angles.X))
	}
	if bits&client.MoveAngle2 > 0 {
		msg.WriteShort(uint16(to.Angles.Y))
	}
	if bits&client.MoveAngle3 > 0 {
		msg.WriteShort(uint16(to.Angles.Z))
	}
	if bits&client.MoveForward > 0 {
		msg.WriteShort(uint16(to.Forward))
	}
	if bits&client.MoveSide > 0 {
		msg.WriteShort(uint16(to.Sideways))
	}
	if bits&client.MoveUp > 0 {
		msg.WriteShort(uint16(to.Up))
	}
	if bits&client.MoveButtons > 0 {
		msg.WriteByte(buttons)
	}
	if bits&client.MoveImpulse > 0 {
		msg.WriteByte(to.Impulse)
	}
	msg.WriteByte(to.Msec)
}
*/

// ParsePacket will parse all the messages in a particular server packet.
func (p *Buffer) ParsePacket(oldFrames map[int32]*pb.Frame) (*pb.Packet, error) {
	if p.Index == p.Length {
		return nil, nil
	}
	out := &pb.Packet{}
	for p.Index < len(p.Data) {
		cmd := p.ReadByte()
		switch cmd {
		case SVCServerData:
			out.ServerData = p.ParseServerData()
		case SVCConfigString:
			out.ConfigStrings = append(out.ConfigStrings, p.ParseConfigString())
		case SVCSpawnBaseline:
			bitmask := p.ParseEntityBitmask()
			number := p.ParseEntityNumber(bitmask)
			out.Baselines = append(out.Baselines, p.ParseEntity(nil, number, bitmask))
		case SVCStuffText:
			out.Stuffs = append(out.Stuffs, p.ParseStuffText())
		case SVCFrame: // includes playerstate and packetentities
			out.Frames = append(out.Frames, p.ParseFrame(oldFrames))
		case SVCPrint:
			out.Prints = append(out.Prints, p.ParsePrint())
		case SVCMuzzleFlash:
			out.MuzzleFlashes = append(out.MuzzleFlashes, p.ParseMuzzleFlash())
		case SVCTempEntity:
			out.TempEnts = append(out.TempEnts, p.ParseTempEntity())
		case SVCLayout:
			out.Layouts = append(out.Layouts, p.ParseLayout())
		case SVCSound:
			out.Sounds = append(out.Sounds, p.ParseSound())
		case SVCCenterPrint:
			out.Centerprints = append(out.Centerprints, p.ParseCenterPrint())
		}
	}
	return out, nil
}

// Write a ServerData proto back to binary
func MarshalServerData(s *pb.ServerInfo) Buffer {
	b := Buffer{}
	b.WriteByte(SVCServerData)
	b.WriteLong(int32(s.GetProtocol()))
	b.WriteLong(int32(s.GetServerCount()))
	if s.GetDemo() {
		b.WriteByte(1)
	} else {
		b.WriteByte(0)
	}
	b.WriteString(s.GetGameDir())
	b.WriteShort(uint16(s.GetClientNumber()))
	b.WriteString(s.GetMapName())
	return b
}

// Write a Configstring proto back to binary
func MarshalConfigstring(cs *pb.ConfigString) Buffer {
	b := Buffer{}
	b.WriteByte(SVCConfigString)
	b.WriteShort(uint16(cs.GetIndex()))
	b.WriteString(cs.GetData())
	return b
}

// Write a StuffText proto back to binary
func MarshalStuffText(st *pb.StuffText) Buffer {
	b := Buffer{}
	b.WriteString(st.GetData())
	return b
}

// Write a Print proto back to binary
func MarshalPrint(p *pb.Print) Buffer {
	b := Buffer{}
	b.WriteByte(byte(p.GetLevel()))
	b.WriteString(p.GetData())
	return b
}

// Write a Muzzleflash proto back to binary
func MarshalFlash(mf *pb.MuzzleFlash) Buffer {
	b := Buffer{}
	b.WriteShort(uint16(mf.GetEntity()))
	b.WriteByte(byte(mf.GetWeapon()))
	return b
}

// Write a TempEntity proto back to binary
func MarshalTempEntity(te *pb.TemporaryEntity) Buffer {
	b := Buffer{}
	b.WriteByte(byte(te.GetType()))
	switch te.GetType() {
	case TentBlood:
		fallthrough
	case TentGunshot:
		fallthrough
	case TentSparks:
		fallthrough
	case TentBulletSparks:
		fallthrough
	case TentScreenSparks:
		fallthrough
	case TentShieldSparks:
		fallthrough
	case TentShotgun:
		fallthrough
	case TentBlaster:
		fallthrough
	case TentGreenBlood:
		fallthrough
	case TentBlaster2:
		fallthrough
	case TentFlechette:
		fallthrough
	case TentHeatBeamSparks:
		fallthrough
	case TentHeatBeamSteam:
		fallthrough
	case TentMoreBlood:
		fallthrough
	case TentElectricSparks:
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteByte(byte(te.GetDirection()))
	case TentSplash:
		fallthrough
	case TentLaserSparks:
		fallthrough
	case TentWeldingSparks:
		fallthrough
	case TentTunnelSparks:
		b.WriteByte(byte(te.GetCount()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteByte(byte(te.GetDirection()))
		b.WriteByte(byte(te.GetColor()))
	case TentBlueHyperBlaster:
		fallthrough
	case TentRailTrail:
		fallthrough
	case TentBubbleTrail:
		fallthrough
	case TentDebugTrail:
		fallthrough
	case TentBubbleTrail2:
		fallthrough
	case TentBFGLaser:
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteCoord(uint16(te.GetPosition2X()))
		b.WriteCoord(uint16(te.GetPosition2Y()))
		b.WriteCoord(uint16(te.GetPosition2Z()))
	case TentGrenadeExplosion:
		fallthrough
	case TentGrenadeExplosionWater:
		fallthrough
	case TentExplosion2:
		fallthrough
	case TentPlasmaExplosion:
		fallthrough
	case TentRocketExplosion:
		fallthrough
	case TentRocketExplosionWater:
		fallthrough
	case TentExplosion1:
		fallthrough
	case TentExplosion1NP:
		fallthrough
	case TentExplosion1Big:
		fallthrough
	case TentBFGExplosion:
		fallthrough
	case TentBFGBigExplosion:
		fallthrough
	case TentBossTeleport:
		fallthrough
	case TentPlainExplosion:
		fallthrough
	case TentChainFistSmoke:
		fallthrough
	case TentTrackerExplosion:
		fallthrough
	case TentTeleportEffect:
		fallthrough
	case TentDBallGoal:
		fallthrough
	case TentWidowSplash:
		fallthrough
	case TentNukeBlast:
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
	case TentParasiteAttack:
		fallthrough
	case TentMedicCableAttack:
		fallthrough
	case TentHeatBeam:
		fallthrough
	case TentMonsterHeatBeam:
		b.WriteShort(uint16(te.GetEntity1()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteCoord(uint16(te.GetPosition2X()))
		b.WriteCoord(uint16(te.GetPosition2Y()))
		b.WriteCoord(uint16(te.GetPosition2Z()))
		b.WriteCoord(uint16(te.GetOffsetX()))
		b.WriteCoord(uint16(te.GetOffsetY()))
		b.WriteCoord(uint16(te.GetOffsetZ()))
	case TentGrappleCable:
		b.WriteShort(uint16(te.GetEntity1()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteCoord(uint16(te.GetPosition2X()))
		b.WriteCoord(uint16(te.GetPosition2Y()))
		b.WriteCoord(uint16(te.GetPosition2Z()))
		b.WriteCoord(uint16(te.GetOffsetX()))
		b.WriteCoord(uint16(te.GetOffsetY()))
		b.WriteCoord(uint16(te.GetOffsetZ()))
	case TentLightning:
		b.WriteShort(uint16(te.GetEntity1()))
		b.WriteShort(uint16(te.GetEntity2()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteCoord(uint16(te.GetPosition2X()))
		b.WriteCoord(uint16(te.GetPosition2Y()))
		b.WriteCoord(uint16(te.GetPosition2Z()))
	case TentFlashlight:
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteShort(uint16(te.GetEntity1()))
	case TentForceWall:
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteCoord(uint16(te.GetPosition2X()))
		b.WriteCoord(uint16(te.GetPosition2Y()))
		b.WriteCoord(uint16(te.GetPosition2Z()))
		b.WriteByte(byte(te.GetColor()))
	case TentSteam:
		b.WriteShort(uint16(te.GetEntity1()))
		b.WriteByte(byte(te.GetCount()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
		b.WriteByte(byte(te.GetDirection()))
		b.WriteByte(byte(te.GetColor()))
		b.WriteShort(uint16(te.GetEntity2()))
		if int32(te.Entity1) != -1 {
			b.WriteLong(int32(te.GetTime()))
		}
	case TentWidowBeamOut:
		b.WriteShort(uint16(te.GetEntity1()))
		b.WriteCoord(uint16(te.GetPosition1X()))
		b.WriteCoord(uint16(te.GetPosition1Y()))
		b.WriteCoord(uint16(te.GetPosition1Z()))
	}
	return b
}

// Write a Layout proto back to binary
func MarshalLayout(lo *pb.Layout) Buffer {
	b := Buffer{}
	b.WriteString(lo.GetData())
	return b
}

// Write a PackedSound proto back to binary
func MarshalSound(s *pb.PackedSound) Buffer {
	b := Buffer{}
	b.WriteByte(byte(s.GetFlags()))
	b.WriteByte(byte(s.GetIndex()))
	if (s.GetFlags() & SoundVolume) > 0 {
		b.WriteByte(byte(s.GetVolume()))
	}
	if (s.GetFlags() & SoundAttenuation) > 0 {
		b.WriteByte(byte(s.GetAttenuation()))
	}
	if (s.GetFlags() & SoundOffset) > 0 {
		b.WriteByte(byte(s.GetTimeOffset()))
	}
	if (s.GetFlags() & SoundEntity) > 0 {
		b.WriteShort(uint16((s.GetEntity() << 3) + s.GetChannel()))
	}
	if (s.GetFlags() & SoundPosition) > 0 {
		b.WriteCoord(uint16(s.GetPositionX()))
		b.WriteCoord(uint16(s.GetPositionY()))
		b.WriteCoord(uint16(s.GetPositionZ()))
	}
	return b
}

// Write a Centerprint proto back to binary
func MarshalCenterPrint(cp *pb.CenterPrint) Buffer {
	b := Buffer{}
	b.WriteString(cp.GetData())
	return b
}

// Write a Frame proto (including playerstate and entities) back to binary
func MarshalFrame(fr *pb.Frame) Buffer {
	msg := Buffer{}
	msg.WriteByte(SVCFrame)
	msg.WriteLong(fr.Number)
	msg.WriteLong(fr.Delta)
	msg.WriteByte(byte(fr.Suppressed))
	msg.WriteByte(byte(fr.AreaBytes))
	for _, ab := range fr.AreaBits {
		msg.WriteByte(byte(ab))
	}

	// from state is empty
	msg.Append(WriteDeltaPlayerstate(nil, fr.PlayerState))

	msg.WriteByte(SVCPacketEntities)
	for _, ent := range fr.GetEntities() {
		msg.Append(WriteDeltaEntity(&pb.PackedEntity{}, ent))
	}
	msg.WriteShort(0) // EoE

	// player-based muzzle flashes
	for _, flash := range fr.GetFlashes1() {
		msg.WriteByte(SVCMuzzleFlash)
		msg.Append(MarshalFlash(flash))
	}
	// monster-based muzzle flashes
	for _, flash := range fr.GetFlashes2() {
		msg.WriteByte(SVCMuzzleFlash2)
		msg.Append(MarshalFlash(flash))
	}
	for _, ent := range fr.GetTemporaryEntities() {
		msg.WriteByte(SVCTempEntity)
		msg.Append(MarshalTempEntity(ent))
	}
	for _, layout := range fr.GetLayouts() {
		msg.WriteByte(SVCLayout)
		msg.Append(MarshalLayout(layout))
	}
	for _, sound := range fr.GetSounds() {
		msg.WriteByte(SVCSound)
		msg.Append(MarshalSound(sound))
	}
	for _, print := range fr.GetPrints() {
		msg.WriteByte(SVCPrint)
		msg.Append(MarshalPrint(print))
	}
	for _, stuff := range fr.GetStufftexts() {
		msg.WriteByte(SVCStuffText)
		msg.Append(MarshalStuffText(stuff))
	}
	for _, cs := range fr.GetConfigstrings() {
		msg.Append(MarshalConfigstring(cs))
	}
	for _, cp := range fr.GetCenterprints() {
		msg.WriteByte(SVCCenterPrint)
		msg.Append(MarshalCenterPrint(cp))
	}
	return msg
}
