package message

import (
	pb "github.com/packetflinger/libq2/proto"
)

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
	var i uint32
	for i = 0; i < MaxStats; i++ {
		if to.GetStats()[i] != from.GetStats()[i] {
			statbits |= 1 << i
		}
	}
	msg.WriteLong(int32(statbits))
	for i = 0; i < MaxStats; i++ {
		if (bits & (1 << i)) > 0 {
			msg.WriteShort(uint16(to.GetStats()[i]))
		}
	}
}
