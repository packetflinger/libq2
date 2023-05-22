package message

import "bytes"

// server to client message types
const (
	SVCBad = iota
	SVCMuzzleFlash
	SVCMuzzleFlash2
	SVCTempEntity
	SVCLayout
	SVCInventory
	SVCNOP
	SVCDisconnect
	SVCReconnect
	SVCSound
	SVCPrint
	SVCStuffText
	SVCServerData
	SVCConfigString
	SVCSpawnBaseline
	SVCCenterPrint
	SVCDownload
	SVCPlayerInfo
	SVCPacketEntities
	SVCDeltaPacketEntities
	SVCFrame
	SVCZPacket   // r1q2
	SVCZDownload // r1q2
	SVCGameState // r1q2/q2pro
	SVCSetting   // r1q2/q2pro
	SVCNumTypes  // r1q2/q2pro
)

// entity state flags
const (
	EntityOrigin1   = 1 << 0
	EntityOrigin2   = 1 << 1
	EntityAngle2    = 1 << 2
	EntityAngle3    = 1 << 3
	EntityFrame8    = 1 << 4
	EntityEvent     = 1 << 5
	EntityRemove    = 1 << 6
	EntityMoreBits1 = 1 << 7

	EntityNumber16  = 1 << 8
	EntityOrigin3   = 1 << 9
	EntityAngle1    = 1 << 10
	EntityModel     = 1 << 11
	EntityRenderFX8 = 1 << 12
	EntityAngle16   = 1 << 13
	EntityEffects8  = 1 << 14
	EntityMoreBits2 = 1 << 15

	EntitySkin8      = 1 << 16
	EntityFrame16    = 1 << 17
	EntityRenderFX16 = 1 << 18
	EntityEffects16  = 1 << 19
	EntityModel2     = 1 << 20
	EntityModel3     = 1 << 21
	EntityModel4     = 1 << 22
	EntityMoreBits3  = 1 << 23

	EntityOldOrigin = 1 << 24
	EntitySkin16    = 1 << 25
	EntitySound     = 1 << 26
	EntitySolid     = 1 << 27
)

// playerstate flags
const (
	PlayerType        = 1 << 0
	PlayerOrigin      = 1 << 1
	PlayerVelocity    = 1 << 2
	PlayerTime        = 1 << 3
	PlayerFlags       = 1 << 4
	PlayerGravity     = 1 << 5
	PlayerDeltaAngles = 1 << 6
	PlayerViewOffset  = 1 << 7

	PlayerViewAngles  = 1 << 8
	PlayerKickAngles  = 1 << 9
	PlayerBlend       = 1 << 10
	PlayerFOV         = 1 << 11
	PlayerWeaponIndex = 1 << 12
	PlayerWeaponFrame = 1 << 13
	PlayerRDFlags     = 1 << 14
	PlayerReserved    = 1 << 15

	PlayerBits = 16
	PlayerMask = (1 << PlayerBits) - 1
)

// Sound properties
const (
	SoundVolume      = 1 << 0 // 1 byte
	SoundAttenuation = 1 << 1 // 1 byte
	SoundPosition    = 1 << 2 // 3 coordinates
	SoundEntity      = 1 << 3 // short 0-2: channel, 3-12: entity
	SoundOffset      = 1 << 4 // 1 byte, msec offset from frame start
)

// temporary entity types
const (
	TentGunshot = iota
	TentBlood
	TentBlaster
	TentRailTrail
	TentShotgun
	TentExplosion1
	TentExplosion2
	TentRocketExplosion
	TentGrenadeExplosion
	TentSparks
	TentSplash
	TentBubbleTrail
	TentScreenSparks
	TentShieldSparks
	TentBulletSparks
	TentLaserSparks
	TentParasiteAttack
	TentRocketExplosionWater
	TentGrenadeExplosionWater
	TentMedicCableAttack
	TentBFGExplosion
	TentBFGBigExplosion
	TentBossTeleport
	TentBFGLaser
	TentGrappleCable
	TentWeldingSparks
	TentGreenBlood
	TentBlueHyperBlaster
	TentPlasmaExplosion
	TentTunnelSparks
	TentBlaster2
	TentRailTrail2
	TentFlame
	TentLightning
	TentDebugTrail
	TentPlainExplosion
	TentFlashlight
	TentForceWall
	TentHeatBeam
	TentMonsterHeatBeam
	TentSteam
	TentBubbleTrail2
	TentMoreBlood
	TentHeatBeamSparks
	TentHeatBeamSteam
	TentChainFistSmoke
	TentElectricSparks
	TentTrackerExplosion
	TentTeleportEffect
	TentDBallGoal
	TentWidowBeamOut
	TentNukeBlast
	TentWidowSplash
	TentExplosion1Big
	TentExplosion1NP
	TentFlechette
	TentNumEntities
)

// configstrings
const (
	CSMapname = 33
)

const (
	RFFrameLerp = 64
	RFBeam      = 128
)

type MessageBuffer struct {
	Buffer []byte
	Index  int
	Length int // maybe not needed
}

func NewMessageBuffer(data []byte) MessageBuffer {
	return MessageBuffer{
		Buffer: data,
		Index:  0,
		Length: len(data),
	}
}

func (m *MessageBuffer) Reset() {
	m.Buffer = make([]byte, len(m.Buffer))
	m.Index = 0
	m.Length = 0
}

// 4 bytes signed
func (msg *MessageBuffer) ReadLong() int32 {
	l := int32(msg.Buffer[msg.Index])
	l += int32(msg.Buffer[msg.Index+1]) << 8
	l += int32(msg.Buffer[msg.Index+2]) << 16
	l += int32(msg.Buffer[msg.Index+3]) << 24
	msg.Index += 4
	return l
}

func (msg *MessageBuffer) WriteLong(data int32) {
	b := []byte{
		byte(data & 0xff),
		byte((data >> 8) & 0xff),
		byte((data >> 16) & 0xff),
		byte((data >> 24) & 0xff),
	}
	msg.Buffer = append(msg.Buffer, b...)
	msg.Index += 4
}

// just grab a subsection of the buffer
func (msg *MessageBuffer) ReadData(length int) []byte {
	start := msg.Index
	msg.Index += length
	return msg.Buffer[start:msg.Index]
}

func (msg *MessageBuffer) WriteData(data []byte) {
	msg.Buffer = append(msg.Buffer, data...)
	msg.Index += len(data)
}

// Keep building a string until we hit a null
func (msg *MessageBuffer) ReadString() string {
	var buffer bytes.Buffer

	// find the next null (terminates the string)
	for i := 0; msg.Buffer[msg.Index] != 0; i++ {
		// we hit the end without finding a null
		if msg.Index == len(msg.Buffer) {
			break
		}

		buffer.WriteString(string(msg.Buffer[msg.Index]))
		msg.Index++
	}

	msg.Index++
	return buffer.String()
}

// Strings are null terminated, so add a 0x00 at the end.
func (msg *MessageBuffer) WriteString(s string) {
	for _, ch := range s {
		msg.WriteByte(byte(ch))
	}

	msg.WriteByte(0)
}

// 2 bytes unsigned
func (msg *MessageBuffer) ReadShort() uint16 {
	s := uint16(msg.Buffer[msg.Index] & 0xff)
	s += uint16(msg.Buffer[msg.Index+1]) << 8
	msg.Index += 2

	return s
}

func (msg *MessageBuffer) WriteShort(s uint16) {
	b := []byte{
		byte(s & 0xff),
		byte((s >> 8) & 0xff),
	}
	msg.Buffer = append(msg.Buffer, b...)
	msg.Index += 2
}

// for consistency
func (msg *MessageBuffer) ReadByte() byte {
	val := byte(msg.Buffer[msg.Index])
	msg.Index++
	return val
}

func (msg *MessageBuffer) WriteByte(b byte) {
	bb := []byte{b}
	msg.Buffer = append(msg.Buffer, bb...)
	msg.Index++
}

// 1 byte signed
func (msg *MessageBuffer) ReadChar() int8 {
	val := int8(msg.Buffer[msg.Index])
	msg.Index++
	return val
}

func (msg *MessageBuffer) WriteChar(c uint8) {
	bb := []byte{byte(c)}
	msg.Buffer = append(msg.Buffer, bb...)
	msg.Index++
}

// 2 bytes signed
func (msg *MessageBuffer) ReadWord() int16 {
	s := int16(msg.Buffer[msg.Index] & 0xff)
	s += int16(msg.Buffer[msg.Index+1]) << 8
	msg.Index += 2

	return s
}

func (msg *MessageBuffer) WriteWord(w int16) {
	b := []byte{
		byte(w & 0xff),
		byte(w >> 8),
	}
	msg.Buffer = append(msg.Buffer, b...)
	msg.Index += 2
}

func (msg *MessageBuffer) ReadCoord() uint16 {
	return msg.ReadShort()
}

func (msg *MessageBuffer) WriteCoord(c uint16) {
	msg.WriteShort(c)
}

func (msg *MessageBuffer) ReadPosition() [3]uint16 {
	x := msg.ReadCoord()
	y := msg.ReadCoord()
	z := msg.ReadCoord()
	return [3]uint16{x, y, z}
}

func (msg *MessageBuffer) ReadDirection() uint8 {
	return msg.ReadByte()
}
