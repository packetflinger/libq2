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

type Parser interface {
	ApplyPacket(packet *pb.Packet)
}

/*
// function pointers for each message type
type Callback struct {
	// message specific callbacks
	ServerData   func(*ServerData)
	ConfigString func(*ConfigString)
	Baseline     func(*PackedEntity)
	Frame        func(*FrameMsg)
	PlayerState  func(*PackedPlayer)
	Entity       func([]*PackedEntity)
	Print        func(*Print)
	Stuff        func(*StuffText)
	Layout       func(*Layout)
	CenterPrint  func(*CenterPrint)
	Sound        func(*PackedSound)
	TempEnt      func(*TemporaryEntity)
	Flash1       func(*MuzzleFlash)
	Flash2       func(*MuzzleFlash)

	// event specific callbacks
	OnConnect func()        // connection to gameserver made
	OnEnter   func()        // you entered the game (begin)
	PreSend   func(*Buffer) //

	// needed for parsing compressed things like playerstates and
	// packetentities
	FrameMap map[int]ServerFrame
}
*/

/*
type ServerFrame struct {
	Server         ServerData
	Frame          FrameMsg
	DeltaFrame     *ServerFrame
	Playerstate    PackedPlayer
	Entities       map[int]PackedEntity
	Baselines      map[int]PackedEntity
	Strings        []ConfigString
	Prints         []Print
	Stuffs         []StuffText
	Layouts        []Layout
	Centerprinters []CenterPrint
	Sounds         []PackedSound
	TempEntities   []TemporaryEntity
	Flash1         []MuzzleFlash
	Flash2         []MuzzleFlash
}
*/

/*
// This "header" is present for both demos and live connections.
// Before being spawned in, the server will send a serverdata message,
// then all the current configstrings, and known entities and their
// attributes at that time.
//
// After this, the server will start sending frames to the client
type GamestateHeader struct {
	Serverdata    ServerData
	Configstrings []ConfigString
	Baselines     []PackedEntity
}
*/

/*
// always the first message received from server
type ServerData struct {
	Protocol     int32
	ServerCount  int32
	Demo         int8
	GameDir      string
	ClientNumber int16
	MapName      string
}

type ConfigString struct {
	Index  uint16
	String string
}

type StuffText struct {
	String string
}

type PackedEntity struct {
	Number      uint32
	Origin      [3]int16
	Angles      [3]int16
	OldOrigin   [3]int16
	ModelIndex  uint8
	ModelIndex2 uint8
	ModelIndex3 uint8
	ModelIndex4 uint8
	SkinNum     uint32
	Effects     uint32
	RenderFX    uint32
	Solid       uint32
	Frame       uint16
	Sound       uint8
	Event       uint8
	Remove      bool
}

type PlayerMoveState struct {
	Type        uint8
	Origin      [3]int16
	Velocity    [3]int16
	Flags       byte
	Time        byte
	Gravity     int16
	DeltaAngles [3]int16
}

type PackedPlayer struct {
	PlayerMove PlayerMoveState
	ViewAngles [3]int16
	ViewOffset [3]int8
	KickAngles [3]int8
	GunAngles  [3]int8
	GunOffset  [3]int8
	GunIndex   uint8
	GunFrame   uint8
	Blend      [4]uint8
	FOV        uint8
	RDFlags    uint8
	Stats      [32]int16
}

type FrameMsg struct {
	Number     int32
	Delta      int32
	Suppressed int8
	AreaBytes  int8
	AreaBits   []byte
}

type Print struct {
	Level  uint8
	String string
}

type PackedSound struct {
	Flags       uint8
	Index       uint8
	Volume      uint8
	Attenuation uint8
	TimeOffset  uint8
	Channel     uint16
	Entity      uint16
	Position    [3]uint16
}

type TemporaryEntity struct {
	Type      uint8
	Position1 [3]uint16
	Position2 [3]uint16
	Offset    [3]uint16
	Direction uint8
	Count     uint8
	Color     uint8
	Entity1   int16
	Entity2   int16
	Time      int32
}

type MuzzleFlash struct {
	Entity uint16
	Weapon uint8
}

type Layout struct {
	Data string
}

type CenterPrint struct {
	Data string
}
*/

type ChallengeResponse struct {
	Number    int
	Protocols []int // protocols support by server (34=orig, 35=r1q2, 36=q2pro)
}

func (m *Buffer) ParseServerData() *pb.ServerInfo {
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
func (m *Buffer) ParseConfigString() *pb.CString {
	return &pb.CString{
		Index:   uint32(m.ReadShort()),
		String_: m.ReadString(),
	}
}

// A baseline is just a normal entity in its default state, from
// a client's perspective
func (m *Buffer) ParseSpawnBaseline() *pb.PackedEntity {
	bitmask := m.ParseEntityBitmask()
	number := m.ParseEntityNumber(bitmask)
	return m.ParseEntity(&pb.PackedEntity{}, number, bitmask)
}

func (m *Buffer) ParseStuffText() *pb.StuffText {
	return &pb.StuffText{String_: m.ReadString()}
}

func (m *Buffer) ParseFrame(oldFrames map[int32]*pb.Frame) *pb.Frame {
	fr := &pb.Frame{}
	fr.Number = int32(m.ReadLong())
	fr.Delta = int32(m.ReadLong())
	fr.Suppressed = uint32(m.ReadByte())
	fr.AreaBytes = uint32(m.ReadByte())
	areabits := m.ReadData(int(fr.GetAreaBytes()))
	for _, ab := range areabits {
		fr.AreaBits = append(fr.AreaBits, uint32(ab))
	}
	deltaFrame := oldFrames[fr.Delta]
	var ps *pb.PackedPlayer
	if m.ReadByte() == SVCPlayerInfo {
		ps = m.ParseDeltaPlayerstate(deltaFrame.GetPlayerState())
	}
	fr.PlayerState = ps
	if m.ReadByte() == SVCPacketEntities {
		fr.Entities = m.ParsePacketEntities(fr.Entities)
	}
	return fr
}

func (m *Buffer) ParsePrint() *pb.Print {
	return &pb.Print{
		Level:   uint32(m.ReadByte()),
		String_: m.ReadString(),
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

func (m *Buffer) ParseTempEntity() *pb.TemporaryEntity {
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

// A gun fired, nearby clients should see the flash
func (m *Buffer) ParseMuzzleFlash() *pb.MuzzleFlash {
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
	return &pb.Layout{
		String_: m.ReadString(),
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
	return &pb.CenterPrint{
		String_: m.ReadString(),
	}
}

func (m *Buffer) ParseChallenge() (ChallengeResponse, error) {
	cl := m.ReadLong()
	if cl != -1 {
		return ChallengeResponse{}, errors.New("not connectionless message, invalid challenge response")
	}

	tokens := strings.Fields(string(m.ReadString()))
	num, err := strconv.Atoi(tokens[1])
	if err != nil {
		return ChallengeResponse{}, errors.New("invalid challenge response")
	}

	pr := []int{}
	protocols := strings.Split(tokens[2][2:], ",")
	for _, p := range protocols {
		pint, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		pr = append(pr, pint)
	}

	return ChallengeResponse{
		Number:    num,
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

func MarshalConfigstring(cs *pb.CString) Buffer {
	b := Buffer{}
	b.WriteByte(SVCConfigString)
	b.WriteShort(uint16(cs.GetIndex()))
	b.WriteString(cs.GetString_())
	return b
}

func MarshalStuffText(st *pb.StuffText) Buffer {
	b := Buffer{}
	b.WriteString(st.GetString_())
	return b
}

func MarshalPrint(p *pb.Print) Buffer {
	b := Buffer{}
	b.WriteByte(byte(p.GetLevel()))
	b.WriteString(p.GetString_())
	return b
}

func MarshalFlash(mf *pb.MuzzleFlash) Buffer {
	b := Buffer{}
	b.WriteShort(uint16(mf.GetEntity()))
	b.WriteByte(byte(mf.GetWeapon()))
	return b
}

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

func MarshalLayout(lo *pb.Layout) Buffer {
	b := Buffer{}
	b.WriteString(lo.GetString_())
	return b
}

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
		b.WriteShort(uint16(s.GetEntity()<<3 + s.GetChannel()))
	}
	if (s.GetFlags() & SoundPosition) > 0 {
		b.WriteCoord(uint16(s.GetPositionX()))
		b.WriteCoord(uint16(s.GetPositionY()))
		b.WriteCoord(uint16(s.GetPositionZ()))
	}
	return b
}

func MarshalCenterPrint(cp *pb.CenterPrint) Buffer {
	b := Buffer{}
	b.WriteString(cp.GetString_())
	return b
}

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
	WriteDeltaPlayer(&pb.PackedPlayer{}, fr.PlayerState, &msg)

	msg.WriteByte(SVCPacketEntities)
	for _, ent := range fr.GetEntities() {
		WriteDeltaEntity(&pb.PackedEntity{}, ent, &msg)
	}
	msg.WriteShort(0) // EoE

	// player-based muzzle flashes
	for _, flash := range fr.GetFlashes1() {
		tmp := MarshalFlash(flash)
		msg.WriteByte(SVCMuzzleFlash)
		msg.Append(tmp)
	}
	// monster-basd muzzle flashes
	for _, flash := range fr.GetFlashes2() {
		tmp := MarshalFlash(flash)
		msg.WriteByte(SVCMuzzleFlash2)
		msg.Append(tmp)
	}
	for _, ent := range fr.GetTemporaryEntities() {
		tmp := MarshalTempEntity(ent)
		msg.WriteByte(SVCTempEntity)
		msg.Append(tmp)
	}
	for _, layout := range fr.GetLayouts() {
		tmp := MarshalLayout(layout)
		msg.WriteByte(SVCLayout)
		msg.Append(tmp)
	}
	for _, sound := range fr.GetSounds() {
		tmp := MarshalSound(sound)
		msg.WriteByte(SVCSound)
		msg.Append(tmp)
	}
	for _, print := range fr.GetPrints() {
		tmp := MarshalPrint(print)
		msg.WriteByte(SVCPrint)
		msg.Append(tmp)
	}
	for _, stuff := range fr.GetStufftexts() {
		tmp := MarshalStuffText(stuff)
		msg.WriteByte(SVCStuffText)
		msg.Append(tmp)
	}
	for _, cs := range fr.GetConfigstrings() {
		tmp := MarshalConfigstring(cs)
		msg.Append(tmp)
	}
	for _, cp := range fr.GetCenterprints() {
		tmp := MarshalCenterPrint(cp)
		msg.WriteByte(SVCCenterPrint)
		msg.Append(tmp)
	}
	return msg
}
