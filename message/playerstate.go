package message

import (
	pb "github.com/packetflinger/libq2/proto"
	"google.golang.org/protobuf/proto"
)

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

// DeltaPlayerBitmask will return a bitmask representing the difference between
// two playerstates. This way only differences are transmitted from server to
// client to save bandwidth/processing since playerstates are emitted on every
// frame (every 100ms by default, and down to every 20ms for q2pro on protocol
// 36). This is used for both parsing and writing playerstates.
//
// The bitmask is an unsigned short
func DeltaPlayerBitmask(from *pb.PackedPlayer, to *pb.PackedPlayer) int {
	mask := int(0)
	mf := from.GetMovestate()
	mt := to.GetMovestate()

	if mf.GetType() != mt.GetType() {
		mask |= PlayerType
	}

	if mf.GetOriginX() != mt.GetOriginX() || mf.GetOriginY() != mt.GetOriginY() || mf.GetOriginZ() != mt.GetOriginZ() {
		mask |= PlayerOrigin
	}
	if mf.GetVelocityX() != mt.GetVelocityX() || mf.GetVelocityY() != mt.GetVelocityY() || mf.GetVelocityZ() != mt.GetVelocityZ() {
		mask |= PlayerVelocity
	}
	if mf.GetTime() != mt.GetTime() {
		mask |= PlayerTime
	}
	if mf.GetFlags() != mt.GetFlags() {
		mask |= PlayerFlags
	}
	if mf.GetGravity() != mt.GetGravity() {
		mask |= PlayerGravity
	}
	if mf.GetDeltaAngleX() != mt.GetDeltaAngleX() || mf.GetDeltaAngleY() != mt.GetDeltaAngleY() || mf.GetDeltaAngleZ() != mt.GetDeltaAngleZ() {
		mask |= PlayerDeltaAngles
	}
	if from.GetViewOffsetX() != to.GetViewOffsetX() || from.GetViewOffsetY() != to.GetViewOffsetY() || from.GetViewOffsetZ() != to.GetViewOffsetZ() {
		mask |= PlayerViewOffset
	}
	if from.GetViewAnglesX() != to.GetViewAnglesX() || from.GetViewAnglesY() != to.GetViewAnglesY() || from.GetViewAnglesZ() != to.GetViewAnglesZ() {
		mask |= PlayerViewAngles
	}
	if from.GetKickAnglesX() != to.GetKickAnglesX() || from.GetKickAnglesY() != to.GetKickAnglesY() || from.GetKickAnglesZ() != to.GetKickAnglesZ() {
		mask |= PlayerKickAngles
	}
	if from.GetBlendW() != to.GetBlendW() || from.GetBlendX() != to.GetBlendX() || from.GetBlendY() != to.GetBlendY() || from.GetBlendZ() != to.GetBlendZ() {
		mask |= PlayerBlend
	}
	if from.GetFov() != to.GetFov() {
		mask |= PlayerFOV
	}
	if from.GetRdFlags() != to.GetRdFlags() {
		mask |= PlayerRDFlags
	}
	if from.GetGunFrame() != to.GetGunFrame() || from.GetGunOffsetX() != to.GetGunOffsetX() || from.GetGunOffsetY() != to.GetGunOffsetY() || from.GetGunOffsetZ() != to.GetGunOffsetZ() || from.GetGunAnglesX() != to.GetGunAnglesX() || from.GetGunAnglesY() != to.GetGunAnglesY() || from.GetGunAnglesZ() != to.GetGunAnglesZ() {
		mask |= PlayerWeaponFrame
	}
	if from.GetGunIndex() != to.GetGunIndex() {
		mask |= PlayerWeaponIndex
	}
	return mask
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
	if m.Index == m.Length { // end of buffer (or empty buffer)
		return nil
	}
	if from != nil {
		to = proto.Clone(from).(*pb.PackedPlayer)
		// proto.Clone() isn't deep, sub-protos need to be copied manually
		pm = proto.Clone(from.Movestate).(*pb.PlayerMove)
		if pm == nil {
			// from might not have playermove defined
			pm = &pb.PlayerMove{}
		}
		for k, v := range from.GetStats() {
			to.Stats[k] = v
		}
	}
	mask := m.ReadWord()

	if (mask & PlayerType) != 0 {
		pm.Type = uint32(m.ReadByte())
	}

	if (mask & PlayerOrigin) != 0 {
		pm.OriginX = int32(m.ReadShort())
		pm.OriginY = int32(m.ReadShort())
		pm.OriginZ = int32(m.ReadShort())
	}

	if (mask & PlayerVelocity) != 0 {
		pm.VelocityX = uint32(m.ReadShort())
		pm.VelocityY = uint32(m.ReadShort())
		pm.VelocityZ = uint32(m.ReadShort())
	}

	if (mask & PlayerTime) != 0 {
		pm.Time = uint32(m.ReadByte())
	}

	if (mask & PlayerFlags) != 0 {
		pm.Flags = uint32(m.ReadByte())
	}

	if (mask & PlayerGravity) != 0 {
		pm.Gravity = int32(m.ReadShort())
	}

	if (mask & PlayerDeltaAngles) != 0 {
		pm.DeltaAngleX = int32(m.ReadShort())
		pm.DeltaAngleY = int32(m.ReadShort())
		pm.DeltaAngleZ = int32(m.ReadShort())
	}

	if (mask & PlayerViewOffset) != 0 {
		to.ViewOffsetX = int32(m.ReadChar())
		to.ViewOffsetY = int32(m.ReadChar())
		to.ViewOffsetZ = int32(m.ReadChar())
	}

	if (mask & PlayerViewAngles) != 0 {
		to.ViewAnglesX = int32(m.ReadShort())
		to.ViewAnglesY = int32(m.ReadShort())
		to.ViewAnglesZ = int32(m.ReadShort())
	}

	if (mask & PlayerKickAngles) != 0 {
		to.KickAnglesX = int32(m.ReadChar())
		to.KickAnglesY = int32(m.ReadChar())
		to.KickAnglesZ = int32(m.ReadChar())
	}

	if (mask & PlayerWeaponIndex) != 0 {
		to.GunIndex = uint32(m.ReadByte())
	}

	if (mask & PlayerWeaponFrame) != 0 {
		to.GunFrame = uint32(m.ReadByte())
		to.GunOffsetX = int32(m.ReadChar())
		to.GunOffsetY = int32(m.ReadChar())
		to.GunOffsetZ = int32(m.ReadChar())
		to.GunAnglesX = int32(m.ReadChar())
		to.GunAnglesY = int32(m.ReadChar())
		to.GunAnglesZ = int32(m.ReadChar())
	}

	if (mask & PlayerBlend) != 0 {
		to.BlendW = int32(m.ReadChar())
		to.BlendX = int32(m.ReadChar())
		to.BlendY = int32(m.ReadChar())
		to.BlendZ = int32(m.ReadChar())
	}

	if (mask & PlayerFOV) != 0 {
		to.Fov = uint32(m.ReadByte())
	}

	if (mask & PlayerRDFlags) != 0 {
		to.RdFlags = uint32(m.ReadByte())
	}

	statsMask := int32(m.ReadLong())
	var i uint32
	for i = 0; i < MaxStats; i++ {
		if (statsMask & (1 << i)) != 0 {
			stats[i] = int32(m.ReadShort())
		}
	}
	to.Stats = stats
	to.Movestate = pm
	return to
}

// WriteDeltaPlayerstate will convert the changes between the `from` and `to`
// playerstates from a textproto to binary that q2 clients understand.
func WriteDeltaPlayerstate(from *pb.PackedPlayer, to *pb.PackedPlayer) Buffer {
	b := Buffer{}
	mask := DeltaPlayerBitmask(from, to)
	b.WriteByte(SVCPlayerInfo)
	b.WriteShort(mask)

	if (mask & PlayerType) > 0 {
		b.WriteByte(int(to.GetMovestate().GetType()))
	}

	if (mask & PlayerOrigin) > 0 {
		b.WriteShort(int(to.GetMovestate().GetOriginX()))
		b.WriteShort(int(to.GetMovestate().GetOriginY()))
		b.WriteShort(int(to.GetMovestate().GetOriginZ()))
	}

	if (mask & PlayerVelocity) > 0 {
		b.WriteShort(int(to.GetMovestate().GetVelocityX()))
		b.WriteShort(int(to.GetMovestate().GetVelocityY()))
		b.WriteShort(int(to.GetMovestate().GetVelocityZ()))
	}

	if (mask & PlayerTime) > 0 {
		b.WriteByte(int(to.GetMovestate().GetTime()))
	}

	if (mask & PlayerFlags) > 0 {
		b.WriteByte(int(to.GetMovestate().GetFlags()))
	}

	if (mask & PlayerGravity) > 0 {
		b.WriteShort(int(to.GetMovestate().GetGravity()))
	}

	if (mask & PlayerDeltaAngles) > 0 {
		b.WriteShort(int(to.GetMovestate().GetDeltaAngleX()))
		b.WriteShort(int(to.GetMovestate().GetDeltaAngleY()))
		b.WriteShort(int(to.GetMovestate().GetDeltaAngleZ()))
	}

	if (mask & PlayerViewOffset) > 0 {
		b.WriteChar(int(to.GetViewOffsetX()))
		b.WriteChar(int(to.GetViewOffsetY()))
		b.WriteChar(int(to.GetViewOffsetZ()))
	}

	if (mask & PlayerViewAngles) > 0 {
		b.WriteShort(int(to.GetViewAnglesX()))
		b.WriteShort(int(to.GetViewAnglesY()))
		b.WriteShort(int(to.GetViewAnglesZ()))
	}

	if (mask & PlayerKickAngles) > 0 {
		b.WriteChar(int(to.GetKickAnglesX()))
		b.WriteChar(int(to.GetKickAnglesY()))
		b.WriteChar(int(to.GetKickAnglesZ()))
	}

	if (mask & PlayerWeaponIndex) > 0 {
		b.WriteByte(int(to.GetGunIndex()))
	}

	if (mask & PlayerWeaponFrame) > 0 {
		b.WriteByte(int(to.GetGunFrame()))
		b.WriteChar(int(to.GetGunOffsetX()))
		b.WriteChar(int(to.GetGunOffsetY()))
		b.WriteChar(int(to.GetGunOffsetZ()))
		b.WriteChar(int(to.GetGunAnglesX()))
		b.WriteChar(int(to.GetGunAnglesY()))
		b.WriteChar(int(to.GetGunAnglesZ()))
	}

	if (mask & PlayerBlend) > 0 {
		b.WriteByte(int(to.GetBlendW()))
		b.WriteByte(int(to.GetBlendX()))
		b.WriteByte(int(to.GetBlendY()))
		b.WriteByte(int(to.GetBlendZ()))
	}

	if (mask & PlayerFOV) > 0 {
		b.WriteByte(int(to.GetFov()))
	}

	if (mask & PlayerRDFlags) > 0 {
		b.WriteByte(int(to.GetRdFlags()))
	}

	statsMask := uint32(0)
	toStats := to.GetStats()
	fromStats := from.GetStats()
	var i uint32
	for i = 0; i < MaxStats; i++ {
		if toStats[i] != fromStats[i] {
			statsMask |= (1 << i)
		}
	}

	b.WriteLong(int(statsMask))
	for i = 0; i < MaxStats; i++ {
		if (statsMask & (1 << i)) != 0 {
			b.WriteShort(int(toStats[i]))
		}
	}
	return b
}
