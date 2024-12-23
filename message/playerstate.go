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
