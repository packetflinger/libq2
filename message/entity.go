package message

import (
	"google.golang.org/protobuf/proto"

	pb "github.com/packetflinger/libq2/proto"
)

// entity state flags
const (
	EntityOrigin1    = 1 << 0
	EntityOrigin2    = 1 << 1
	EntityAngle2     = 1 << 2
	EntityAngle3     = 1 << 3
	EntityFrame8     = 1 << 4
	EntityEvent      = 1 << 5
	EntityRemove     = 1 << 6
	EntityMoreBits1  = 1 << 7 // keep going
	EntityNumber16   = 1 << 8
	EntityOrigin3    = 1 << 9
	EntityAngle1     = 1 << 10
	EntityModel      = 1 << 11
	EntityRenderFX8  = 1 << 12
	EntityAngle16    = 1 << 13
	EntityEffects8   = 1 << 14
	EntityMoreBits2  = 1 << 15 // keep going
	EntitySkin8      = 1 << 16
	EntityFrame16    = 1 << 17
	EntityRenderFX16 = 1 << 18
	EntityEffects16  = 1 << 19
	EntityModel2     = 1 << 20
	EntityModel3     = 1 << 21
	EntityModel4     = 1 << 22
	EntityMoreBits3  = 1 << 23 // keep going
	EntityOldOrigin  = 1 << 24
	EntitySkin16     = 1 << 25
	EntitySound      = 1 << 26
	EntitySolid      = 1 << 27
	EntityModel16    = 1 << 28
	EntityMoreFX8    = 1 << 29
	EntityAlpha      = 1 << 30
	EntityMoreBits4  = 1 << 31 // keep going
	EntityScale      = 1 << 32
	EntityMoreFX16   = 1 << 33

	EntitySkin32     = EntitySkin8 | EntitySkin16
	EntityEffects32  = EntityEffects8 | EntityEffects16
	EntityRenderFX32 = EntityRenderFX8 | EntityRenderFX16
	EntityMoreFX32   = EntityMoreFX8 | EntityMoreFX16
)

// Read up to the first 4 bytes of an entity, depending on the
// previous ones. This value tells you what data is in the rest
// of the entity message.
func (m *Buffer) ParseEntityBitmask() uint32 {
	if m.Index == m.Length {
		return 0
	}
	mask := uint32(m.ReadByte())
	if (mask & EntityMoreBits1) != 0 {
		mask |= (uint32(m.ReadByte()) << 8)
	}
	if (mask & EntityMoreBits2) != 0 {
		mask |= (uint32(m.ReadByte()) << 16)
	}
	if (mask & EntityMoreBits3) != 0 {
		mask |= (uint32(m.ReadByte()) << 24)
	}
	return mask
}

// ParseEntityNumber will read the edict number of an entity. This number will
// be between 1 and MaxEntities. Entity 0 is the world.
func (m *Buffer) ParseEntityNumber(flags uint32) uint16 {
	if m.Index == m.Length {
		return 0
	}
	num := uint16(0)
	if (flags & EntityNumber16) != 0 {
		num = uint16(m.ReadShort())
	} else {
		num = uint16(m.ReadByte())
	}
	return num
}

// ParseEntity will parse an entity from stream and return an uncompressed
// PackedEntity proto. It uses the `from` param to decompress, this acts as a
// baseline, applies the changes and returns a clone of that full PackedEntity.
func (m *Buffer) ParseEntity(from *pb.PackedEntity, num uint16, bits uint32) *pb.PackedEntity {
	if m.Index == m.Length {
		return nil
	}
	to := &pb.PackedEntity{}
	if from != nil {
		to = proto.Clone(from).(*pb.PackedEntity)
	}
	to.Number = uint32(num)

	if bits == 0 {
		return to
	}

	if (bits & EntityModel) != 0 {
		to.ModelIndex = uint32(m.ReadByte())
	}

	if (bits & EntityModel2) != 0 {
		to.ModelIndex2 = uint32(m.ReadByte())
	}

	if (bits & EntityModel3) != 0 {
		to.ModelIndex3 = uint32(m.ReadByte())
	}

	if (bits & EntityModel4) != 0 {
		to.ModelIndex4 = uint32(m.ReadByte())
	}

	if (bits & EntityFrame8) != 0 {
		to.Frame = uint32(m.ReadByte())
	}

	if (bits & EntityFrame16) != 0 {
		to.Frame = uint32(m.ReadShort())
	}

	if (bits & (EntitySkin8 | EntitySkin16)) == (EntitySkin8 | EntitySkin16) {
		to.Skin = uint32(m.ReadLong())
	} else if (bits & EntitySkin8) != 0 {
		to.Skin = uint32(m.ReadByte())
	} else if (bits & EntitySkin16) != 0 {
		to.Skin = uint32(m.ReadWord())
	}

	if (bits & (EntityEffects8 | EntityEffects16)) == (EntityEffects8 | EntityEffects16) {
		to.Effects = uint32(m.ReadLong())
	} else if (bits & EntityEffects8) != 0 {
		to.Effects = uint32(m.ReadByte())
	} else if (bits & EntityEffects16) != 0 {
		to.Effects = uint32(m.ReadWord())
	}

	if (bits & (EntityRenderFX8 | EntityRenderFX16)) == (EntityRenderFX8 | EntityRenderFX16) {
		to.RenderFx = uint32(m.ReadLong())
	} else if (bits & EntityRenderFX8) != 0 {
		to.RenderFx = uint32(m.ReadByte())
	} else if (bits & EntityRenderFX16) != 0 {
		to.RenderFx = uint32(m.ReadWord())
	}

	if (bits & EntityOrigin1) != 0 {
		to.OriginX = int32(m.ReadShort())
	}

	if (bits & EntityOrigin2) != 0 {
		to.OriginY = int32(m.ReadShort())
	}

	if (bits & EntityOrigin3) != 0 {
		to.OriginZ = int32(m.ReadShort())
	}

	if (bits & EntityAngle1) != 0 {
		to.AngleX = int32(m.ReadByte())
	}

	if (bits & EntityAngle2) != 0 {
		to.AngleY = int32(m.ReadByte())
	}

	if (bits & EntityAngle3) != 0 {
		to.AngleZ = int32(m.ReadByte())
	}

	if (bits & EntityOldOrigin) != 0 {
		to.OldOriginX = int32(m.ReadShort())
		to.OldOriginY = int32(m.ReadShort())
		to.OldOriginZ = int32(m.ReadShort())
	}

	if (bits & EntitySound) != 0 {
		to.Sound = uint32(m.ReadByte())
	}

	if (bits & EntityEvent) != 0 {
		to.Event = uint32(m.ReadByte())
	}

	if (bits & EntitySolid) != 0 {
		to.Solid = uint32(m.ReadWord())
	}

	if (bits & EntityRemove) != 0 {
		to.Remove = true
	}
	return to
}

// ParsePacketEntities will parse an `SVC_PACKETENTITIES` msg. This is the
// last of the 3-tuple of msgs sent from the server for each frame. There is no
// delimiter between entities, once the entity is fully parsed it immediately
// moves on to the next. If the entity number is 0, the end of the clod of ents
// has been reached.
//
// This has to decompress entities as they're parsed or else there is no way to
// tell new values from existing. This makes it impossible to decompress all
// entities after the fact.
func (m *Buffer) ParsePacketEntities(from map[int32]*pb.PackedEntity) map[int32]*pb.PackedEntity {
	if m.Index == m.Length {
		return nil
	}
	out := make(map[int32]*pb.PackedEntity)
	for k := range from {
		out[k] = proto.Clone(from[k]).(*pb.PackedEntity)
	}
	for {
		bits := m.ParseEntityBitmask()
		num := m.ParseEntityNumber(bits)
		if num <= 0 {
			break
		}
		orig, ok := out[int32(num)]
		if !ok {
			orig = &pb.PackedEntity{}
		}
		out[int32(num)] = m.ParseEntity(orig, num, bits)
	}
	return out
}

// WriteDeltaEntity will emit the differences between `from` and `to` as binary
// that q2 clients can understand.
func WriteDeltaEntity(from *pb.PackedEntity, to *pb.PackedEntity) Buffer {
	b := Buffer{}
	bits := DeltaEntityBitmask(to, from)

	// write the bitmask first
	b.WriteByte(bits & 255)
	if (bits & 0xff000000) > 0 {
		b.WriteByte((bits >> 8) & 255)
		b.WriteByte((bits >> 16) & 255)
		b.WriteByte((bits >> 24) & 255)
	} else if (bits & 0x00ff0000) > 0 {
		b.WriteByte((bits >> 8) & 255)
		b.WriteByte((bits >> 16) & 255)
	} else if (bits & 0x0000ff00) > 0 {
		b.WriteByte((bits >> 8) & 255)
	}

	// write the edict number
	if (bits & EntityNumber16) > 0 {
		b.WriteShort(int(to.GetNumber()))
	} else {
		b.WriteByte(int(to.GetNumber()))
	}

	if (bits & EntityModel) > 0 {
		b.WriteByte(int(to.GetModelIndex()))
	}

	if (bits & EntityModel2) > 0 {
		b.WriteByte(int(to.GetModelIndex2()))
	}

	if (bits & EntityModel3) > 0 {
		b.WriteByte(int(to.GetModelIndex3()))
	}

	if (bits & EntityModel4) > 0 {
		b.WriteByte(int(to.GetModelIndex4()))
	}

	if (bits & EntityFrame8) > 0 {
		b.WriteByte(int(to.GetFrame()))
	} else if (bits & EntityFrame16) > 0 {
		b.WriteShort(int(to.GetFrame()))
	}

	if (bits & (EntitySkin8 | EntitySkin16)) == (EntitySkin8 | EntitySkin16) {
		b.WriteLong(int(to.GetSkin()))
	} else if (bits & EntitySkin8) > 0 {
		b.WriteByte(int(to.GetSkin()))
	} else if (bits & EntitySkin16) > 0 {
		b.WriteShort(int(to.GetSkin()))
	}

	if (bits & (EntityEffects8 | EntityEffects16)) == (EntityEffects8 | EntityEffects16) {
		b.WriteLong(int(to.GetEffects()))
	} else if (bits & EntityEffects8) > 0 {
		b.WriteByte(int(to.GetEffects()))
	} else if (bits & EntityEffects16) > 0 {
		b.WriteShort(int(to.GetEffects()))
	}

	if (bits & (EntityRenderFX8 | EntityRenderFX16)) == (EntityRenderFX8 | EntityRenderFX16) {
		b.WriteLong(int(to.GetRenderFx()))
	} else if (bits & EntityRenderFX8) > 0 {
		b.WriteByte(int(to.GetRenderFx()))
	} else if (bits & EntityRenderFX16) > 0 {
		b.WriteShort(int(to.GetRenderFx()))
	}

	if (bits & EntityOrigin1) > 0 {
		b.WriteShort(int(to.GetOriginX()))
	}

	if (bits & EntityOrigin2) > 0 {
		b.WriteShort(int(to.GetOriginY()))
	}

	if (bits & EntityOrigin3) > 0 {
		b.WriteShort(int(to.GetOriginZ()))
	}

	if (bits & EntityAngle1) > 0 {
		b.WriteByte(int(to.GetAngleX()))
	}

	if (bits & EntityAngle2) > 0 {
		b.WriteByte(int(to.GetAngleY()))
	}

	if (bits & EntityAngle3) > 0 {
		b.WriteByte(int(to.GetAngleZ()))
	}

	if (bits & EntityOldOrigin) > 0 {
		b.WriteShort(int(to.GetOldOriginX()))
		b.WriteShort(int(to.GetOldOriginY()))
		b.WriteShort(int(to.GetOldOriginZ()))
	}

	if (bits & EntitySound) > 0 {
		b.WriteByte(int(to.GetSound()))
	}

	if (bits & EntityEvent) > 0 {
		b.WriteByte(int(to.GetEvent()))
	}

	if (bits & EntitySolid) > 0 {
		b.WriteShort(int(to.GetSolid()))
	}
	return b
}

// DeltaEntityBitmask will return the bitmask representing the differences
// between the `to` and `from` entities.
func DeltaEntityBitmask(to *pb.PackedEntity, from *pb.PackedEntity) int {
	bits := int(0)
	mask := int(0xffff8000)
	if to == nil {
		to = &pb.PackedEntity{}
	}
	if from == nil {
		from = &pb.PackedEntity{}
	}

	if to.GetRemove() {
		bits |= EntityRemove
	}

	if to.GetOriginX() != from.GetOriginX() {
		bits |= EntityOrigin1
	}

	if to.GetOriginY() != from.GetOriginY() {
		bits |= EntityOrigin2
	}

	if to.GetOriginZ() != from.GetOriginZ() {
		bits |= EntityOrigin3
	}

	if to.GetAngleX() != from.GetAngleX() {
		bits |= EntityAngle1
	}

	if to.GetAngleY() != from.GetAngleY() {
		bits |= EntityAngle2
	}

	if to.GetAngleZ() != from.GetAngleZ() {
		bits |= EntityAngle3
	}

	if to.GetSkin() != from.GetSkin() {
		if (int(to.GetSkin()) & mask) > 0 {
			bits |= EntitySkin8 | EntitySkin16
		} else if (to.GetSkin() & uint32(0x0000ff00)) > 0 {
			bits |= EntitySkin16
		} else {
			bits |= EntitySkin8
		}
	}

	if to.GetFrame() != from.GetFrame() {
		if (uint16(to.GetFrame()) & uint16(0xff00)) > 0 {
			bits |= EntityFrame16
		} else {
			bits |= EntityFrame8
		}
	}

	if to.Effects != from.Effects {
		if (int(to.Effects) & mask) > 0 {
			bits |= EntityEffects8 | EntityEffects16
		} else if (to.Effects & 0x0000ff00) > 0 {
			bits |= EntityEffects16
		} else {
			bits |= EntityEffects8
		}
	}

	if to.GetRenderFx() != from.GetRenderFx() {
		if (int(to.GetRenderFx()) & mask) > 0 {
			bits |= EntityRenderFX8 | EntityRenderFX16
		} else if (to.GetRenderFx() & 0x0000ff00) > 0 {
			bits |= EntityRenderFX16
		} else {
			bits |= EntityRenderFX8
		}
	}

	if to.GetSolid() != from.GetSolid() {
		bits |= EntitySolid
	}

	if to.GetEvent() != from.GetEvent() {
		bits |= EntityEvent
	}

	if to.GetModelIndex() != from.GetModelIndex() {
		bits |= EntityModel
	}

	if to.GetModelIndex2() != from.GetModelIndex2() {
		bits |= EntityModel2
	}

	if to.GetModelIndex3() != from.GetModelIndex3() {
		bits |= EntityModel3
	}

	if to.GetModelIndex4() != from.GetModelIndex4() {
		bits |= EntityModel4
	}

	if to.GetSound() != from.GetSound() {
		bits |= EntitySound
	}

	if (to.GetRenderFx() & RFFrameLerp) > 0 {
		bits |= EntityOldOrigin
	} else if (to.GetRenderFx() & RFBeam) > 0 {
		bits |= EntityOldOrigin
	}

	if (to.GetNumber() & 0xff00) > 0 {
		bits |= EntityNumber16
	}

	if (bits & 0xff000000) > 0 {
		bits |= EntityMoreBits3 | EntityMoreBits2 | EntityMoreBits1
	} else if (bits & 0x00ff0000) > 0 {
		bits |= EntityMoreBits2 | EntityMoreBits1
	} else if (bits & 0x0000ff00) > 0 {
		bits |= EntityMoreBits1
	}

	return bits
}
