package message

import (
	"bytes"
	"encoding/binary"

	"github.com/packetflinger/libq2/util"
)

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

// Sound properties
const (
	SoundVolume      = 1 << 0 // 1 byte
	SoundAttenuation = 1 << 1 // 1 byte
	SoundPosition    = 1 << 2 // 3 coordinates
	SoundEntity      = 1 << 3 // short 0-2: channel, 3-12: entity
	SoundOffset      = 1 << 4 // 1 byte, msec offset from frame start
	SoundIndex16     = 1 << 5 // index is 16 bits
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

type Buffer struct {
	Data   []byte
	Index  int
	Length int // maybe not needed
}

func NewBuffer(data []byte) Buffer {
	return Buffer{
		Data:   data,
		Index:  0,
		Length: len(data),
	}
}

// Reinitialize the back to zero
func (m *Buffer) Reset() {
	m.Data = []byte{}
	m.Index = 0
	m.Length = 0
}

// combine 2 buffers, set index to the end
func (m *Buffer) Append(m2 Buffer) {
	m.Data = append(m.Data, m2.Data...)
	m.Index = len(m.Data)
	m.Length = len(m.Data)
}

// Set the internal pointer to a specific value between zero and the langth of
// the buffer.
func (m *Buffer) Seek(offset int) {
	off := util.Clamp(offset, 0, len(m.Data))
	m.Index = off
}

// Get the length of the current buffer
func (m *Buffer) Size() int {
	return len(m.Data)
}

// Set the internal pointer back to the beginning of the buffer. Same as `seek(0)`
func (m *Buffer) Rewind() {
	m.Index = 0
}

// Find out if the buffer has any data or not
func (m *Buffer) IsEmpty() bool {
	return m.Length == 0
}

// As the buffer been parsed all the way to the end?
func (m *Buffer) AtEnd() bool {
	return m.Index == m.Length
}

// Read 4 bytes and construct a 32 bit signed integer from the data. The `Uint32`
// from the binary package returns an unsigned `int32“, it then needs to be
// casted to a `int32` and then finally to an `int`. Skipping that intermediate
// cast results in the wrong value.
func (msg *Buffer) ReadLong() int {
	if msg.Index+4 > msg.Length {
		return 0
	}
	return int(int32(binary.LittleEndian.Uint32(msg.ReadData(4))))
}

// 4 bytes signed
func (msg *Buffer) WriteLong(data int) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(int32(data)))
	msg.WriteData(b)
}

// just grab a subsection of the buffer
func (msg *Buffer) ReadData(length int) []byte {
	if msg.Index+length > msg.Length {
		return []byte{}
	}
	start := msg.Index
	msg.Index += length
	return msg.Data[start:msg.Index]
}

func (msg *Buffer) WriteData(data []byte) {
	msg.Data = append(msg.Data, data...)
	msg.Index += len(data)
}

// Keep building a string until we hit a null
func (msg *Buffer) ReadString() string {
	var buffer bytes.Buffer

	if msg.IsEmpty() || msg.AtEnd() {
		return ""
	}

	// find the next null (terminates the string)
	for i := 0; msg.Data[msg.Index] != 0; i++ {
		buffer.WriteString(string(msg.Data[msg.Index]))
		msg.Index++

		// we hit the end without finding a null
		if msg.Index == len(msg.Data) {
			break
		}
	}
	msg.Index++
	return buffer.String()
}

// Strings are null terminated, so add a 0x00 at the end.
func (msg *Buffer) WriteString(s string) {
	for _, ch := range s {
		msg.WriteByte(int(ch))
	}
	msg.WriteByte(0)
}

// 2 bytes signed
func (msg *Buffer) ReadShort() int {
	if msg.Index+2 > msg.Length {
		return 0
	}
	return int(int16(binary.LittleEndian.Uint16(msg.ReadData(2))))
}

func (msg *Buffer) WriteShort(s int) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(int16(s)))
	msg.WriteData(b)
}

// unsigned
func (msg *Buffer) ReadByte() int {
	if msg.Index == msg.Length {
		return 0
	}
	val := int(uint8(msg.Data[msg.Index]))
	msg.Index++
	return val
}

func (msg *Buffer) WriteByte(b int) {
	bb := []byte{uint8(int8(b))}
	msg.Data = append(msg.Data, bb...)
	msg.Index++
}

// 1 byte signed
func (msg *Buffer) ReadChar() int {
	if msg.Index == msg.Length {
		return 0
	}
	val := int(int8(msg.Data[msg.Index]))
	msg.Index++
	return val
}

func (msg *Buffer) WriteChar(c int) {
	bb := []byte{uint8(int8(c))}
	msg.Data = append(msg.Data, bb...)
	msg.Index++
}

// 2 bytes unsigned
func (msg *Buffer) ReadWord() int {
	if msg.Index+2 > msg.Length {
		return 0
	}
	return int(binary.LittleEndian.Uint16(msg.ReadData(2)))
}

func (msg *Buffer) WriteWord(w int) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(int16(w)))
	msg.WriteData(b)
}

func (msg *Buffer) ReadCoord() int {
	return msg.ReadShort()
}

func (msg *Buffer) WriteCoord(c int) {
	msg.WriteShort(c)
}

func (msg *Buffer) ReadPosition() [3]int {
	return [3]int{
		msg.ReadCoord(),
		msg.ReadCoord(),
		msg.ReadCoord(),
	}
}

func (msg *Buffer) ReadDirection() int {
	return msg.ReadByte()
}

func (msg *Buffer) ReadVarInt64() uint64 {
	if msg.Index+8 > msg.Length {
		return 0
	}
	var v uint64
	var c, bits int
	for {
		c = int(msg.ReadByte())
		if c == -1 {
			break
		}
		v |= uint64((c & 0x7f) << bits)
		bits += 7
		if ((c & 0x80) == 0) || (bits >= 64) {
			break
		}
	}
	return v
}
