package message

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	util "github.com/packetflinger/libq2/util"
)

const (
	MaxItems         = 256
	MaxStats         = 32
	MaxEntities      = 1024
	MaxConfigStrings = 2080
	CLCMove          = 2
	CLCUserinfo      = 3
	CLCStringCommand = 4
)

// function pointers for each message type
type MessageCallbacks struct {
	ServerDataCB   func(ServerData)
	ConfigStringCB func(ConfigString)
	BaselineCB     func(PackedEntity)
	FrameCB        func(FrameMsg)
	PlayerStateCB  func(PackedPlayer)
	EntityCB       func([]PackedEntity)
	PrintCB        func(Print)
	StuffCB        func(StuffText)
	LayoutCB       func(Layout)
	CenterPrintCB  func(CenterPrint)
	SoundCB        func(PackedSound)
	TempEntCB      func(TemporaryEntity)
	Flash1CB       func(MuzzleFlash)
	Flash2CB       func(MuzzleFlash)
}

type ServerFrame struct {
	Server         ServerData
	Frame          FrameMsg
	Playerstate    PackedPlayer
	Entities       [MaxEntities]PackedEntity
	Baselines      [MaxEntities]PackedEntity
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

type ChallengeResponse struct {
	Number    int
	Protocols []int // protocols support by server (34=orig, 35=r1q2, 36=q2pro)
}

type ClientPacket struct {
	Sequence1   int32
	Sequence2   int32
	QPort       uint16 //
	Reliable1   bool   // requires an ack
	Reliable2   bool
	MessageType byte   // what kind of msg?
	Data        []byte // the actual msg
}

// Parse through a clod of various messages. In the case of demos, these lumps
// will be read from disk, for a live client, they'll be received via the network
// every 0.1 seconds
func ParseMessageLump(buf MessageBuffer, cb MessageCallbacks) (ServerFrame, error) {
	sf := ServerFrame{}

	for buf.Index < len(buf.Buffer) {
		cmd := buf.ReadByte()

		switch cmd {
		case SVCServerData:
			s := buf.ParseServerData()
			sf.Server = s
			if cb.ServerDataCB != nil {
				cb.ServerDataCB(s)
			}

		case SVCConfigString:
			cs := buf.ParseConfigString()
			sf.Strings = append(sf.Strings, cs)
			if cb.ConfigStringCB != nil {
				cb.ConfigStringCB(cs)
			}

		case SVCSpawnBaseline:
			bl := buf.ParseSpawnBaseline()
			sf.Baselines[bl.Number] = bl
			if cb.BaselineCB != nil {
				cb.BaselineCB(bl)
			}

		case SVCStuffText:
			st := buf.ParseStuffText()
			sf.Stuffs = append(sf.Stuffs, st)
			if cb.StuffCB != nil {
				cb.StuffCB(st)
			}

		case SVCFrame:
			fr := buf.ParseFrame()
			sf.Frame = fr
			if cb.FrameCB != nil {
				cb.FrameCB(fr)
			}

		case SVCPlayerInfo:
			ps := buf.ParseDeltaPlayerstate(PackedPlayer{})
			sf.Playerstate = ps
			if cb.PlayerStateCB != nil {
				cb.PlayerStateCB(ps)
			}

		case SVCPacketEntities:
			ents := buf.ParsePacketEntities()
			for _, e := range ents {
				sf.Entities[e.Number] = e
			}
			if cb.EntityCB != nil {
				cb.EntityCB(ents)
			}

		case SVCPrint:
			p := buf.ParsePrint()
			sf.Prints = append(sf.Prints, p)
			if cb.PrintCB != nil {
				cb.PrintCB(p)
			}

		case SVCSound:
			s := buf.ParseSound()
			sf.Sounds = append(sf.Sounds, s)
			if cb.SoundCB != nil {
				cb.SoundCB(s)
			}

		case SVCTempEntity:
			te := buf.ParseTempEntity()
			sf.TempEntities = append(sf.TempEntities, te)
			if cb.TempEntCB != nil {
				cb.TempEntCB(te)
			}

		case SVCMuzzleFlash:
			mf := buf.ParseMuzzleFlash()
			sf.Flash1 = append(sf.Flash1, mf)
			if cb.Flash1CB != nil {
				cb.Flash1CB(mf)
			}

		case SVCMuzzleFlash2:
			mf := buf.ParseMuzzleFlash()
			if cb.Flash2CB != nil {
				cb.Flash2CB(mf)
			}

		case SVCLayout:
			l := buf.ParseLayout()
			if cb.LayoutCB != nil {
				cb.LayoutCB(l)
			}

		case SVCInventory:
			// nobody cares about inventory msgs, just parse them in case they're present
			buf.ParseInventory()

		case SVCCenterPrint:
			c := buf.ParseCenterPrint()
			if cb.CenterPrintCB != nil {
				cb.CenterPrintCB(c)
			}

		case SVCBad:
			continue

		default:
			return ServerFrame{}, fmt.Errorf("unknown CMD: %d\n%s", cmd, hex.Dump(buf.Buffer[buf.Index-1:]))
		}
	}
	return sf, nil
}

func (m *MessageBuffer) ParseServerData() ServerData {
	return ServerData{
		Protocol:     m.ReadLong(),
		ServerCount:  m.ReadLong(),
		Demo:         int8(m.ReadByte()),
		GameDir:      m.ReadString(),
		ClientNumber: int16(m.ReadShort()),
		MapName:      m.ReadString(),
	}
}

// Configstrings are strings sent to each client and associated
// with an index. They're referenced by index in various playces
// such as layouts, etc.
func (m *MessageBuffer) ParseConfigString() ConfigString {
	return ConfigString{
		Index:  m.ReadShort(),
		String: m.ReadString(),
	}
}

// A baseline is just a normal entity in its default state, from
// a client's perspective
func (m *MessageBuffer) ParseSpawnBaseline() PackedEntity {
	bitmask := m.ParseEntityBitmask()
	number := m.ParseEntityNumber(bitmask)
	return m.ParseEntity(PackedEntity{}, number, bitmask)
}

// Read up to the first 4 bytes of an entity, depending on the
// previous ones. This value tells you what data is in the rest
// of the entity message.
func (m *MessageBuffer) ParseEntityBitmask() uint32 {
	bits := uint32(m.ReadByte())

	if bits&EntityMoreBits1 != 0 {
		bits |= (uint32(m.ReadByte()) << 8)
	}

	if bits&EntityMoreBits2 != 0 {
		bits |= (uint32(m.ReadByte()) << 16)
	}

	if bits&EntityMoreBits3 != 0 {
		bits |= (uint32(m.ReadByte()) << 24)
	}

	return uint32(bits)
}

func (m *MessageBuffer) ParseEntityNumber(flags uint32) uint16 {
	num := uint16(0)
	if flags&EntityNumber16 != 0 {
		num = uint16(m.ReadShort())
	} else {
		num = uint16(m.ReadByte())
	}

	return num
}

func (m *MessageBuffer) ParseEntity(from PackedEntity, num uint16, bits uint32) PackedEntity {
	to := from
	to.Number = uint32(num)

	if bits == 0 {
		return to
	}

	if bits&EntityModel != 0 {
		to.ModelIndex = uint8(m.ReadByte())
	}

	if bits&EntityModel2 != 0 {
		to.ModelIndex2 = uint8(m.ReadByte())
	}

	if bits&EntityModel3 != 0 {
		to.ModelIndex3 = uint8(m.ReadByte())
	}

	if bits&EntityModel4 != 0 {
		to.ModelIndex4 = uint8(m.ReadByte())
	}

	if bits&EntityFrame8 != 0 {
		to.Frame = uint16(m.ReadByte())
	}

	if bits&EntityFrame16 != 0 {
		to.Frame = uint16(m.ReadShort())
	}

	if (bits & (EntitySkin8 | EntitySkin16)) == (EntitySkin8 | EntitySkin16) {
		to.SkinNum = uint32(m.ReadLong())
	} else if bits&EntitySkin8 != 0 {
		to.SkinNum = uint32(m.ReadByte())
	} else if bits&EntitySkin16 != 0 {
		to.SkinNum = uint32(m.ReadWord())
	}

	if (bits & (EntityEffects8 | EntityEffects16)) == (EntityEffects8 | EntityEffects16) {
		to.Effects = uint32(m.ReadLong())
	} else if bits&EntityEffects8 != 0 {
		to.Effects = uint32(m.ReadByte())
	} else if bits&EntityEffects16 != 0 {
		to.Effects = uint32(m.ReadWord())
	}

	if (bits & (EntityRenderFX8 | EntityRenderFX16)) == (EntityRenderFX8 | EntityRenderFX16) {
		to.RenderFX = uint32(m.ReadLong())
	} else if bits&EntityRenderFX8 != 0 {
		to.RenderFX = uint32(m.ReadByte())
	} else if bits&EntityRenderFX16 != 0 {
		to.RenderFX = uint32(m.ReadWord())
	}

	if bits&EntityOrigin1 != 0 {
		to.Origin[0] = int16(m.ReadShort())
	}

	if bits&EntityOrigin2 != 0 {
		to.Origin[1] = int16(m.ReadShort())
	}

	if bits&EntityOrigin3 != 0 {
		to.Origin[2] = int16(m.ReadShort())
	}

	if bits&EntityAngle1 != 0 {
		to.Angles[0] = int16(m.ReadByte())
	}

	if bits&EntityAngle2 != 0 {
		to.Angles[1] = int16(m.ReadByte())
	}

	if bits&EntityAngle3 != 0 {
		to.Angles[2] = int16(m.ReadByte())
	}

	if bits&EntityOldOrigin != 0 {
		to.OldOrigin[0] = int16(m.ReadShort())
		to.OldOrigin[1] = int16(m.ReadShort())
		to.OldOrigin[2] = int16(m.ReadShort())
	}

	if bits&EntitySound != 0 {
		to.Sound = uint8(m.ReadByte())
	}

	if bits&EntityEvent != 0 {
		to.Event = uint8(m.ReadByte())
	}

	if bits&EntitySolid != 0 {
		to.Solid = uint32(m.ReadWord())
	}

	return to
}

func (m *MessageBuffer) ParseStuffText() StuffText {
	return StuffText{String: m.ReadString()}
}

func (m *MessageBuffer) ParseFrame() FrameMsg {
	N := int32(m.ReadLong())
	D := int32(m.ReadLong())
	S := int8(m.ReadByte())
	A := int8(m.ReadByte())
	Ab := m.ReadData(int(A))

	return FrameMsg{
		Number:     N,
		Delta:      D,
		Suppressed: S,
		AreaBytes:  A,
		AreaBits:   Ab,
	}
}

func (m *MessageBuffer) ParseDeltaPlayerstate(ps PackedPlayer) PackedPlayer {
	bits := m.ReadWord()
	pm := PlayerMoveState{}

	if bits&PlayerType != 0 {
		pm.Type = uint8(m.ReadByte())
	}

	if bits&PlayerOrigin != 0 {
		pm.Origin[0] = int16(m.ReadShort())
		pm.Origin[1] = int16(m.ReadShort())
		pm.Origin[2] = int16(m.ReadShort())
	}

	if bits&PlayerVelocity != 0 {
		pm.Velocity[0] = int16(m.ReadShort())
		pm.Velocity[1] = int16(m.ReadShort())
		pm.Velocity[2] = int16(m.ReadShort())
	}

	if bits&PlayerTime != 0 {
		pm.Time = m.ReadByte()
	}

	if bits&PlayerFlags != 0 {
		pm.Flags = m.ReadByte()
	}

	if bits&PlayerGravity != 0 {
		pm.Gravity = int16(m.ReadShort())
	}

	if bits&PlayerDeltaAngles != 0 {
		pm.DeltaAngles[0] = int16(m.ReadShort())
		pm.DeltaAngles[1] = int16(m.ReadShort())
		pm.DeltaAngles[2] = int16(m.ReadShort())
	}

	if bits&PlayerViewOffset != 0 {
		ps.ViewOffset[0] = int8(m.ReadChar())
		ps.ViewOffset[1] = int8(m.ReadChar())
		ps.ViewOffset[2] = int8(m.ReadChar())
	}

	if bits&PlayerViewAngles != 0 {
		ps.ViewAngles[0] = int16(m.ReadShort())
		ps.ViewAngles[1] = int16(m.ReadShort())
		ps.ViewAngles[2] = int16(m.ReadShort())
	}

	if bits&PlayerKickAngles != 0 {
		ps.KickAngles[0] = int8(m.ReadChar())
		ps.KickAngles[1] = int8(m.ReadChar())
		ps.KickAngles[2] = int8(m.ReadChar())
	}

	if bits&PlayerWeaponIndex != 0 {
		ps.GunIndex = uint8(m.ReadByte())
	}

	if bits&PlayerWeaponFrame != 0 {
		ps.GunFrame = uint8(m.ReadByte())
		ps.GunOffset[0] = int8(m.ReadChar())
		ps.GunOffset[1] = int8(m.ReadChar())
		ps.GunOffset[2] = int8(m.ReadChar())
		ps.GunAngles[0] = int8(m.ReadChar())
		ps.GunAngles[1] = int8(m.ReadChar())
		ps.GunAngles[2] = int8(m.ReadChar())
	}

	if bits&PlayerBlend != 0 {
		ps.Blend[0] = uint8(m.ReadChar())
		ps.Blend[1] = uint8(m.ReadChar())
		ps.Blend[2] = uint8(m.ReadChar())
		ps.Blend[3] = uint8(m.ReadChar())
	}

	if bits&PlayerFOV != 0 {
		ps.FOV = uint8(m.ReadByte())
	}

	if bits&PlayerRDFlags != 0 {
		ps.RDFlags = uint8(m.ReadByte())
	}

	statbits := int32(m.ReadLong())
	for i := 0; i < 32; i++ {
		if statbits&(1<<i) != 0 {
			ps.Stats[i] = int16(m.ReadShort())
		}
	}

	ps.PlayerMove = pm
	return ps
}

// A S->C msg containing all entities the client should
// know aobut for a particular frame
func (m *MessageBuffer) ParsePacketEntities() []PackedEntity {
	ents := []PackedEntity{}

	for {
		bits := m.ParseEntityBitmask()
		num := m.ParseEntityNumber(bits)

		if num <= 0 {
			break
		}

		entity := m.ParseEntity(PackedEntity{}, num, bits)
		ents = append(ents, entity)
	}

	return ents
}

func (m *MessageBuffer) ParsePrint() Print {
	return Print{
		Level:  uint8(m.ReadByte()),
		String: m.ReadString(),
	}
}

func (m *MessageBuffer) ParseSound() PackedSound {
	s := PackedSound{}
	s.Flags = m.ReadByte()
	s.Index = m.ReadByte()

	if (s.Flags & SoundVolume) > 0 {
		s.Volume = m.ReadByte()
	} else {
		s.Volume = 1
	}

	if (s.Flags & SoundAttenuation) > 0 {
		s.Attenuation = m.ReadByte()
	} else {
		s.Attenuation = 1
	}

	if (s.Flags & SoundOffset) > 0 {
		s.TimeOffset = m.ReadByte()
	} else {
		s.TimeOffset = 0
	}

	if (s.Flags & SoundEntity) > 0 {
		s.Channel = m.ReadShort() & 7
		s.Entity = s.Channel >> 3
	} else {
		s.Channel = 0
		s.Entity = 0
	}

	if (s.Flags & SoundPosition) > 0 {
		s.Position = m.ReadPosition()
	}

	return s
}

// Figure out the difference between two playerstates
func (to *PackedPlayer) DeltaPlayerstateBitmask(from *PackedPlayer) uint16 {
	bits := uint16(0)

	if to.PlayerMove.Type != from.PlayerMove.Type {
		bits |= PlayerType
	}

	if !util.VectorCompare(to.PlayerMove.Origin, from.PlayerMove.Origin) {
		bits |= PlayerOrigin
	}

	if !util.VectorCompare(to.PlayerMove.Velocity, from.PlayerMove.Velocity) {
		bits |= PlayerVelocity
	}

	if to.PlayerMove.Time != from.PlayerMove.Time {
		bits |= PlayerTime
	}

	if to.PlayerMove.Flags != from.PlayerMove.Flags {
		bits |= PlayerFlags
	}

	if to.PlayerMove.Gravity != from.PlayerMove.Gravity {
		bits |= PlayerTime
	}

	if !util.VectorCompare(to.PlayerMove.DeltaAngles, from.PlayerMove.DeltaAngles) {
		bits |= PlayerDeltaAngles
	}

	if !util.VectorCompare8(to.ViewOffset, from.ViewOffset) {
		bits |= PlayerViewOffset
	}

	if !util.VectorCompare(to.ViewAngles, from.ViewAngles) {
		bits |= PlayerViewAngles
	}

	if !util.VectorCompare8(to.KickAngles, from.KickAngles) {
		bits |= PlayerKickAngles
	}

	if !util.Vector4Compare8(to.Blend, from.Blend) {
		bits |= PlayerBlend
	}

	if to.FOV != from.FOV {
		bits |= PlayerFOV
	}

	if to.RDFlags != from.RDFlags {
		bits |= PlayerRDFlags
	}

	if to.GunFrame != from.GunFrame ||
		!util.VectorCompare8(to.GunOffset, from.GunOffset) ||
		!util.VectorCompare8(to.GunAngles, from.GunAngles) {
		bits |= PlayerWeaponFrame
	}

	if to.GunIndex != from.GunIndex {
		bits |= PlayerWeaponIndex
	}

	return bits
}

// Build a playerstate message, but only the differences between to and from.
func (msg *MessageBuffer) WriteDeltaPlayerstate(to *PackedPlayer, from *PackedPlayer) {
	bits := to.DeltaPlayerstateBitmask(from)
	msg.WriteByte(SVCPlayerInfo)
	msg.WriteShort(bits)

	if bits&PlayerType > 0 {
		msg.WriteByte(to.PlayerMove.Type)
	}

	if bits&PlayerOrigin > 0 {
		msg.WriteShort(uint16(to.PlayerMove.Origin[0]))
		msg.WriteShort(uint16(to.PlayerMove.Origin[1]))
		msg.WriteShort(uint16(to.PlayerMove.Origin[2]))
	}

	if bits&PlayerVelocity > 0 {
		msg.WriteShort(uint16(to.PlayerMove.Velocity[0]))
		msg.WriteShort(uint16(to.PlayerMove.Velocity[1]))
		msg.WriteShort(uint16(to.PlayerMove.Velocity[2]))
	}

	if bits&PlayerTime > 0 {
		msg.WriteByte(to.PlayerMove.Time)
	}

	if bits&PlayerFlags > 0 {
		msg.WriteByte(to.PlayerMove.Flags)
	}

	if bits&PlayerGravity > 0 {
		msg.WriteShort(uint16(to.PlayerMove.Gravity))
	}

	if bits&PlayerDeltaAngles > 0 {
		msg.WriteShort(uint16(to.PlayerMove.DeltaAngles[0]))
		msg.WriteShort(uint16(to.PlayerMove.DeltaAngles[1]))
		msg.WriteShort(uint16(to.PlayerMove.DeltaAngles[2]))
	}

	if bits&PlayerViewOffset > 0 {
		msg.WriteChar(uint8(to.ViewOffset[0]))
		msg.WriteChar(uint8(to.ViewOffset[1]))
		msg.WriteChar(uint8(to.ViewOffset[2]))
	}

	if bits&PlayerViewAngles > 0 {
		msg.WriteShort(uint16(to.ViewAngles[0]))
		msg.WriteShort(uint16(to.ViewAngles[1]))
		msg.WriteShort(uint16(to.ViewAngles[2]))
	}

	if bits&PlayerKickAngles > 0 {
		msg.WriteChar(uint8(to.KickAngles[0]))
		msg.WriteChar(uint8(to.KickAngles[1]))
		msg.WriteChar(uint8(to.KickAngles[2]))
	}

	if bits&PlayerWeaponIndex > 0 {
		msg.WriteByte(to.GunIndex)
	}

	if bits&PlayerWeaponFrame > 0 {
		msg.WriteByte(to.GunFrame)
		msg.WriteChar(uint8(to.GunOffset[0]))
		msg.WriteChar(uint8(to.GunOffset[1]))
		msg.WriteChar(uint8(to.GunOffset[2]))
		msg.WriteChar(uint8(to.GunAngles[0]))
		msg.WriteChar(uint8(to.GunAngles[1]))
		msg.WriteChar(uint8(to.GunAngles[2]))
	}

	if bits&PlayerFOV > 0 {
		msg.WriteByte(to.FOV)
	}

	if bits&PlayerRDFlags > 0 {
		msg.WriteByte(to.RDFlags)
	}

	// compress the stats
	statbits := int32(0)
	for i := 0; i < MaxStats; i++ {
		if to.Stats[i] != from.Stats[i] {
			statbits |= 1 << i
		}
	}

	msg.WriteLong(statbits)
	for i := 0; i < MaxStats; i++ {
		if (statbits & (1 << i)) > 0 {
			msg.WriteShort(uint16(to.Stats[i]))
		}
	}
}

// Find the differences between these two Entities
func (to *PackedEntity) DeltaEntityBitmask(from *PackedEntity) int {
	bits := 0
	mask := uint32(0xffff8000)

	if to.Origin[0] != from.Origin[0] {
		bits |= EntityOrigin1
	}

	if to.Origin[1] != from.Origin[1] {
		bits |= EntityOrigin2
	}

	if to.Origin[2] != from.Origin[2] {
		bits |= EntityOrigin3
	}

	if to.Angles[0] != from.Angles[0] {
		bits |= EntityAngle1
	}

	if to.Angles[1] != from.Angles[1] {
		bits |= EntityAngle2
	}

	if to.Angles[2] != from.Angles[2] {
		bits |= EntityAngle3
	}

	if to.SkinNum != from.SkinNum {
		if to.SkinNum&mask&mask > 0 {
			bits |= EntitySkin8 | EntitySkin16
		} else if to.SkinNum&uint32(0x0000ff00) > 0 {
			bits |= EntitySkin16
		} else {
			bits |= EntitySkin8
		}
	}

	if to.Frame != from.Frame {
		if to.Frame&uint16(0xff00) > 0 {
			bits |= EntityFrame16
		} else {
			bits |= EntityFrame8
		}
	}

	if to.Effects != from.Effects {
		if to.Effects&mask > 0 {
			bits |= EntityEffects8 | EntityEffects16
		} else if to.Effects&0x0000ff00 > 0 {
			bits |= EntityEffects16
		} else {
			bits |= EntityEffects8
		}
	}

	if to.RenderFX != from.RenderFX {
		if to.RenderFX&mask > 0 {
			bits |= EntityRenderFX8 | EntityRenderFX16
		} else if to.RenderFX&0x0000ff00 > 0 {
			bits |= EntityRenderFX16
		} else {
			bits |= EntityRenderFX8
		}
	}

	if to.Solid != from.Solid {
		bits |= EntitySolid
	}

	if to.Event != from.Event {
		bits |= EntityEvent
	}

	if to.ModelIndex != from.ModelIndex {
		bits |= EntityModel
	}

	if to.ModelIndex2 != from.ModelIndex2 {
		bits |= EntityModel2
	}

	if to.ModelIndex3 != from.ModelIndex3 {
		bits |= EntityModel3
	}

	if to.ModelIndex4 != from.ModelIndex4 {
		bits |= EntityModel4
	}

	if to.Sound != from.Sound {
		bits |= EntitySound
	}

	if to.RenderFX&RFFrameLerp > 0 {
		bits |= EntityOldOrigin
	} else if to.RenderFX&RFBeam > 0 {
		bits |= EntityOldOrigin
	}

	if to.Number&0xff00 > 0 {
		bits |= EntityNumber16
	}

	if bits&0xff000000 > 0 {
		bits |= EntityMoreBits3 | EntityMoreBits2 | EntityMoreBits1
	} else if bits&0x00ff0000 > 0 {
		bits |= EntityMoreBits2 | EntityMoreBits1
	} else if bits&0x0000ff00 > 0 {
		bits |= EntityMoreBits1
	}

	return bits
}

// Compare from and to and only write what's different.
// This is "delta compression"
func (m *MessageBuffer) WriteDeltaEntity(from PackedEntity, to PackedEntity) {
	bits := to.DeltaEntityBitmask(&from)

	// write the bitmask first
	m.WriteByte(byte(bits & 255))
	if bits&0xff000000 > 0 {
		m.WriteByte(byte((bits >> 8) & 255))
		m.WriteByte(byte((bits >> 16) & 255))
		m.WriteByte(byte((bits >> 24) & 255))
	} else if bits&0x00ff0000 > 0 {
		m.WriteByte(byte((bits >> 8) & 255))
		m.WriteByte(byte((bits >> 16) & 255))
	} else if bits&0x0000ff00 > 0 {
		m.WriteByte(byte((bits >> 8) & 255))
	}

	// write the edict number
	if bits&EntityNumber16 > 0 {
		m.WriteShort(uint16(to.Number))
	} else {
		m.WriteByte(byte(to.Number))
	}

	if bits&EntityModel > 0 {
		m.WriteByte(to.ModelIndex)
	}

	if bits&EntityModel2 > 0 {
		m.WriteByte(to.ModelIndex2)
	}

	if bits&EntityModel3 > 0 {
		m.WriteByte(to.ModelIndex3)
	}

	if bits&EntityModel4 > 0 {
		m.WriteByte(to.ModelIndex4)
	}

	if bits&EntityFrame8 > 0 {
		m.WriteByte(byte(to.Frame))
	} else if bits&EntityFrame16 > 0 {
		m.WriteShort(to.Frame)
	}

	if (bits & (EntitySkin8 | EntitySkin16)) == (EntitySkin8 | EntitySkin16) {
		m.WriteLong(int32(to.SkinNum))
	} else if bits&EntitySkin8 > 0 {
		m.WriteByte(byte(to.SkinNum))
	} else if bits&EntitySkin16 > 0 {
		m.WriteShort(uint16(to.SkinNum))
	}

	if (bits & (EntityEffects8 | EntityEffects16)) == (EntityEffects8 | EntityEffects16) {
		m.WriteLong(int32(to.Effects))
	} else if bits&EntityEffects8 > 0 {
		m.WriteByte(byte(to.Effects))
	} else if bits&EntityEffects16 > 0 {
		m.WriteShort(uint16(to.Effects))
	}

	if (bits & (EntityRenderFX8 | EntityRenderFX16)) == (EntityRenderFX8 | EntityRenderFX16) {
		m.WriteLong(int32(to.RenderFX))
	} else if bits&EntityRenderFX8 > 0 {
		m.WriteByte(byte(to.RenderFX))
	} else if bits&EntityRenderFX16 > 0 {
		m.WriteShort(uint16(to.RenderFX))
	}

	if bits&EntityOrigin1 > 0 {
		m.WriteShort(uint16(to.Origin[0]))
	}

	if bits&EntityOrigin2 > 0 {
		m.WriteShort(uint16(to.Origin[1]))
	}

	if bits&EntityOrigin3 > 0 {
		m.WriteShort(uint16(to.Origin[2]))
	}

	if bits&EntityAngle1 > 0 {
		m.WriteByte(byte(to.Angles[0] >> 8))
	}

	if bits&EntityAngle2 > 0 {
		m.WriteByte(byte(to.Angles[1] >> 8))
	}

	if bits&EntityAngle3 > 0 {
		m.WriteByte(byte(to.Angles[2] >> 8))
	}

	if bits&EntityOldOrigin > 0 {
		m.WriteShort(uint16(to.OldOrigin[0]))
		m.WriteShort(uint16(to.OldOrigin[1]))
		m.WriteShort(uint16(to.OldOrigin[2]))
	}

	if bits&EntitySound > 0 {
		m.WriteByte(to.Sound)
	}

	if bits&EntityEvent > 0 {
		m.WriteByte(to.Event)
	}

	if bits&EntitySolid > 0 {
		m.WriteShort(uint16(to.Solid))
	}
}

func (m *MessageBuffer) WriteDeltaFrame(from *ServerFrame, to *ServerFrame) {
	m.WriteByte(SVCFrame)
	m.WriteLong(to.Frame.Number)
	m.WriteLong(from.Frame.Number)
	m.WriteByte(byte(to.Frame.Suppressed))
	m.WriteByte(byte(to.Frame.AreaBytes))
	m.WriteData(to.Frame.AreaBits)
}

func (m *MessageBuffer) ParseTempEntity() TemporaryEntity {
	te := TemporaryEntity{}

	te.Type = m.ReadByte()
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
		te.Position1 = m.ReadPosition()
		te.Direction = m.ReadDirection()
	case TentSplash:
		fallthrough
	case TentLaserSparks:
		fallthrough
	case TentWeldingSparks:
		fallthrough
	case TentTunnelSparks:
		te.Count = m.ReadByte()
		te.Position1 = m.ReadPosition()
		te.Direction = m.ReadDirection()
		te.Color = m.ReadByte()
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
		te.Position1 = m.ReadPosition()
		te.Position2 = m.ReadPosition()
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
		te.Position1 = m.ReadPosition()
	case TentParasiteAttack:
		fallthrough
	case TentMedicCableAttack:
		fallthrough
	case TentHeatBeam:
		fallthrough
	case TentMonsterHeatBeam:
		te.Entity1 = int16(m.ReadShort())
		te.Position1 = m.ReadPosition()
		te.Position2 = m.ReadPosition()
		te.Offset = m.ReadPosition()
	case TentGrappleCable:
		te.Entity1 = int16(m.ReadShort())
		te.Position1 = m.ReadPosition()
		te.Position2 = m.ReadPosition()
		te.Offset = m.ReadPosition()
	case TentLightning:
		te.Entity1 = int16(m.ReadShort())
		te.Entity2 = int16(m.ReadShort())
		te.Position1 = m.ReadPosition()
		te.Position2 = m.ReadPosition()
	case TentFlashlight:
		te.Position1 = m.ReadPosition()
		te.Entity1 = int16(m.ReadShort())
	case TentForceWall:
		te.Position1 = m.ReadPosition()
		te.Position2 = m.ReadPosition()
		te.Color = m.ReadByte()
	case TentSteam:
		te.Entity1 = int16(m.ReadShort())
		te.Count = m.ReadByte()
		te.Position1 = m.ReadPosition()
		te.Direction = m.ReadDirection()
		te.Color = m.ReadByte()
		te.Entity2 = int16(m.ReadShort())
		if te.Entity1 != -1 {
			te.Time = m.ReadLong()
		}
	case TentWidowBeamOut:
		te.Entity1 = int16(m.ReadShort())
		te.Position1 = m.ReadPosition()
	}

	return te
}

// A gun fired, nearby clients should see the flash
func (m *MessageBuffer) ParseMuzzleFlash() MuzzleFlash {
	return MuzzleFlash{
		Entity: m.ReadShort(),
		Weapon: m.ReadByte(),
	}
}

// A layout is a string of code to represent how things need
// to be arranged on the screen. The intermission screen
// after a TDM match for example with players, scores, pings,
// stats, etc is an example
func (m *MessageBuffer) ParseLayout() Layout {
	return Layout{
		Data: m.ReadString(),
	}
}

// 2 bytes for every item
func (m *MessageBuffer) ParseInventory() {
	// we don't actually care about this, just parsing it
	inv := [MaxItems]uint16{}
	for i := 0; i < MaxItems; i++ {
		inv[i] = m.ReadShort()
	}
}

// A string that should appear temporarily in the center of the screen
func (m *MessageBuffer) ParseCenterPrint() CenterPrint {
	return CenterPrint{
		Data: m.ReadString(),
	}
}

func (m *MessageBuffer) ParseChallenge() (ChallengeResponse, error) {
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

func NewConnectionlessMessage(str string) MessageBuffer {
	msg := MessageBuffer{}
	msg.WriteLong(-1)
	msg.WriteString(str)
	return msg
}

func NewClientCommand(str string) MessageBuffer {
	msg := MessageBuffer{}
	msg.WriteByte(CLCStringCommand)
	msg.WriteString(str)
	return msg
}

func (p ClientPacket) Marshal() []byte {
	msg := MessageBuffer{}
	msg.WriteLong(p.Sequence1)
	if p.Reliable1 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteLong(p.Sequence2)
	if p.Reliable2 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteShort(uint16(p.QPort))
	msg.WriteByte(p.MessageType)
	msg.WriteData(p.Data)
	return msg.Buffer
}

// Given a sequence number, figure out if it's reliable and if so
// what the actual sequence number is
/*
func ValidateSequence(s uint32) (bool, uint32) {
	if s&0x80000000 > 0 {
		return true, s & 0x7fffffff
	}
	return false, s
}

func SequenceValue(seq uint32, reliable bool) uint32 {
	if reliable {
		return seq | 0x80000000
	}
	return seq
}
*/
