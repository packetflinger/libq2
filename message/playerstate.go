package message

import (
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/proto"
)

// DeltaPlayerBitmask will return a bitmask representing the difference between
// two playerstates. This way only differences are transmitted from server to
// client to save bandwidth/processing since playerstates are emitted on every
// frame (every 100ms by default, and down to every 20ms for q2pro on protocol
// 36).
//
// The bitmask is an unsigned short
func DeltaPlayerBitmask(from *pb.PackedPlayer, to *pb.PackedPlayer) uint16 {
	bits := uint16(0)
	mf := from.GetMovestate()
	mt := to.GetMovestate()

	if mf.GetType() != mt.GetType() {
		bits |= PlayerType
	}

	if mf.GetOriginX() != mt.GetOriginX() || mf.GetOriginY() != mt.GetOriginY() || mf.GetOriginZ() != mt.GetOriginZ() {
		bits |= PlayerOrigin
	}
	if mf.GetVelocityX() != mt.GetVelocityX() || mf.GetVelocityY() != mt.GetVelocityY() || mf.GetVelocityZ() != mt.GetVelocityZ() {
		bits |= PlayerVelocity
	}
	if mf.GetTime() != mt.GetTime() {
		bits |= PlayerTime
	}
	if mf.GetFlags() != mt.GetFlags() {
		bits |= PlayerFlags
	}
	if mf.GetGravity() != mt.GetGravity() {
		bits |= PlayerGravity
	}
	if mf.GetDeltaAngleX() != mt.GetDeltaAngleX() || mf.GetDeltaAngleY() != mt.GetDeltaAngleY() || mf.GetDeltaAngleZ() != mt.GetDeltaAngleZ() {
		bits |= PlayerDeltaAngles
	}
	if from.GetViewOffsetX() != to.GetViewOffsetX() || from.GetViewOffsetY() != to.GetViewOffsetY() || from.GetViewOffsetZ() != to.GetViewOffsetZ() {
		bits |= PlayerViewOffset
	}
	if from.GetViewAnglesX() != to.GetViewAnglesX() || from.GetViewAnglesY() != to.GetViewAnglesY() || from.GetViewAnglesZ() != to.GetViewAnglesZ() {
		bits |= PlayerViewAngles
	}
	if from.GetKickAnglesX() != to.GetKickAnglesX() || from.GetKickAnglesY() != to.GetKickAnglesY() || from.GetKickAnglesZ() != to.GetKickAnglesZ() {
		bits |= PlayerKickAngles
	}
	if from.GetBlendW() != to.GetBlendW() || from.GetBlendX() != to.GetBlendX() || from.GetBlendY() != to.GetBlendY() || from.GetBlendZ() != to.GetBlendZ() {
		bits |= PlayerBlend
	}
	if from.GetFov() != to.GetFov() {
		bits |= PlayerFOV
	}
	if from.GetRdFlags() != to.GetRdFlags() {
		bits |= PlayerRDFlags
	}
	if from.GetGunFrame() != to.GetGunFrame() || from.GetGunOffsetX() != to.GetGunOffsetX() || from.GetGunOffsetY() != to.GetGunOffsetY() || from.GetGunOffsetZ() != to.GetGunOffsetZ() || from.GetGunAnglesX() != to.GetGunAnglesX() || from.GetGunAnglesY() != to.GetGunAnglesY() || from.GetGunAnglesZ() != to.GetGunAnglesZ() {
		bits |= PlayerWeaponFrame
	}
	if from.GetGunIndex() != to.GetGunIndex() {
		bits |= PlayerWeaponIndex
	}
	return bits
}

// ParseDeltaPlayerstate will merge a previous playerstate with one parsed
// from the receiver. Only the changed values are copied.
//
// The merge (decompression) has to happen at parse-time because values can
// change TO zero, and after parsing is complete there is no way to tell if
// a zero is a new value or a missing value.
func (m *Buffer) ParseDeltaPlayerstate(from *pb.PackedPlayer) *pb.PackedPlayer {
	to := &pb.PackedPlayer{}
	pm := &pb.PlayerMove{}
	stats := make(map[uint32]int32)

	if from != nil {
		to = proto.Clone(from).(*pb.PackedPlayer)
		// proto.Clone() isn't deep, sub-protos need to be copied manually
		pm = proto.Clone(from.Movestate).(*pb.PlayerMove)
		for k, v := range from.GetStats() {
			to.Stats[k] = v
		}
	}
	bits := m.ReadWord()

	if bits&PlayerType != 0 {
		pm.Type = uint32(m.ReadByte())
	}

	if bits&PlayerOrigin != 0 {
		pm.OriginX = int32(m.ReadShort())
		pm.OriginY = int32(m.ReadShort())
		pm.OriginZ = int32(m.ReadShort())
	}

	if bits&PlayerVelocity != 0 {
		pm.VelocityX = uint32(m.ReadShort())
		pm.VelocityY = uint32(m.ReadShort())
		pm.VelocityZ = uint32(m.ReadShort())
	}

	if bits&PlayerTime != 0 {
		pm.Time = uint32(m.ReadByte())
	}

	if bits&PlayerFlags != 0 {
		pm.Flags = uint32(m.ReadByte())
	}

	if bits&PlayerGravity != 0 {
		pm.Gravity = int32(m.ReadShort())
	}

	if bits&PlayerDeltaAngles != 0 {
		pm.DeltaAngleX = int32(m.ReadShort())
		pm.DeltaAngleY = int32(m.ReadShort())
		pm.DeltaAngleZ = int32(m.ReadShort())
	}

	if bits&PlayerViewOffset != 0 {
		to.ViewOffsetX = int32(m.ReadChar())
		to.ViewOffsetY = int32(m.ReadChar())
		to.ViewOffsetZ = int32(m.ReadChar())
	}

	if bits&PlayerViewAngles != 0 {
		to.ViewAnglesX = int32(m.ReadShort())
		to.ViewAnglesY = int32(m.ReadShort())
		to.ViewAnglesZ = int32(m.ReadShort())
	}

	if bits&PlayerKickAngles != 0 {
		to.KickAnglesX = int32(m.ReadChar())
		to.KickAnglesY = int32(m.ReadChar())
		to.KickAnglesZ = int32(m.ReadChar())
	}

	if bits&PlayerWeaponIndex != 0 {
		to.GunIndex = uint32(m.ReadByte())
	}

	if bits&PlayerWeaponFrame != 0 {
		to.GunFrame = uint32(m.ReadByte())
		to.GunOffsetX = int32(m.ReadChar())
		to.GunOffsetY = int32(m.ReadChar())
		to.GunOffsetZ = int32(m.ReadChar())
		to.GunAnglesX = int32(m.ReadChar())
		to.GunAnglesY = int32(m.ReadChar())
		to.GunAnglesZ = int32(m.ReadChar())
	}

	if bits&PlayerBlend != 0 {
		to.BlendW = int32(m.ReadChar())
		to.BlendX = int32(m.ReadChar())
		to.BlendY = int32(m.ReadChar())
		to.BlendZ = int32(m.ReadChar())
	}

	if bits&PlayerFOV != 0 {
		to.Fov = uint32(m.ReadByte())
	}

	if bits&PlayerRDFlags != 0 {
		to.RdFlags = uint32(m.ReadByte())
	}

	statbits := int32(m.ReadLong())
	var i uint32
	for i = 0; i < MaxStats; i++ {
		if statbits&(1<<i) != 0 {
			stats[i] = int32(m.ReadShort())
		}
	}
	to.Stats = stats
	to.Movestate = pm
	return to
}

func WriteDeltaPlayer(from *pb.PackedPlayer, to *pb.PackedPlayer, msg *Buffer) {
	bits := DeltaPlayerBitmask(from, to)
	msg.WriteByte(SVCPlayerInfo)
	msg.WriteShort(bits)

	if bits&PlayerType > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetType()))
	}

	if bits&PlayerOrigin > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetOriginX()))
		msg.WriteShort(uint16(to.GetMovestate().GetOriginY()))
		msg.WriteShort(uint16(to.GetMovestate().GetOriginZ()))
	}

	if bits&PlayerVelocity > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityX()))
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityY()))
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityZ()))
	}

	if bits&PlayerTime > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetTime()))
	}

	if bits&PlayerFlags > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetFlags()))
	}

	if bits&PlayerGravity > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetGravity()))
	}

	if bits&PlayerDeltaAngles > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleX()))
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleY()))
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleZ()))
	}

	if bits&PlayerViewOffset > 0 {
		msg.WriteChar(uint8(to.GetViewOffsetX()))
		msg.WriteChar(uint8(to.GetViewOffsetY()))
		msg.WriteChar(uint8(to.GetViewOffsetZ()))
	}

	if bits&PlayerViewAngles > 0 {
		msg.WriteShort(uint16(to.GetViewAnglesX()))
		msg.WriteShort(uint16(to.GetViewAnglesY()))
		msg.WriteShort(uint16(to.GetViewAnglesZ()))
	}

	if bits&PlayerKickAngles > 0 {
		msg.WriteChar(uint8(to.GetKickAnglesX()))
		msg.WriteChar(uint8(to.GetKickAnglesY()))
		msg.WriteChar(uint8(to.GetKickAnglesZ()))
	}

	if bits&PlayerWeaponIndex > 0 {
		msg.WriteByte(byte(to.GetGunIndex()))
	}

	if bits&PlayerWeaponFrame > 0 {
		msg.WriteByte(byte(to.GetGunFrame()))
		msg.WriteChar(uint8(to.GetGunOffsetX()))
		msg.WriteChar(uint8(to.GetGunOffsetY()))
		msg.WriteChar(uint8(to.GetGunOffsetZ()))
		msg.WriteChar(uint8(to.GetGunAnglesX()))
		msg.WriteChar(uint8(to.GetGunAnglesY()))
		msg.WriteChar(uint8(to.GetGunAnglesZ()))
	}

	if bits&PlayerBlend > 0 {
		msg.WriteByte(byte(to.GetBlendW()))
		msg.WriteByte(byte(to.GetBlendX()))
		msg.WriteByte(byte(to.GetBlendY()))
		msg.WriteByte(byte(to.GetBlendZ()))
	}

	if bits&PlayerFOV > 0 {
		msg.WriteByte(byte(to.GetFov()))
	}

	if bits&PlayerRDFlags > 0 {
		msg.WriteByte(byte(to.GetRdFlags()))
	}

	statbits := uint32(0)
	toStats := to.GetStats()
	fromStats := from.GetStats()
	var i uint32
	for i = 0; i < MaxStats; i++ {
		if toStats[i] != fromStats[i] {
			statbits |= 1 << i
		}
	}

	msg.WriteLong(int32(statbits))
	for i = 0; i < MaxStats; i++ {
		if (statbits & (1 << i)) != 0 {
			msg.WriteShort(uint16(toStats[i]))
		}
	}
}
