package message

import (
	"google.golang.org/protobuf/proto"

	pb "github.com/packetflinger/libq2/proto"
)

// Read up to the first 4 bytes of an entity, depending on the
// previous ones. This value tells you what data is in the rest
// of the entity message.
func (m *Buffer) ParseEntityBitmask() uint32 {
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

func (m *Buffer) ParseEntityNumber(flags uint32) uint16 {
	num := uint16(0)
	if flags&EntityNumber16 != 0 {
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
	to := &pb.PackedEntity{}
	if from != nil {
		to = proto.Clone(from).(*pb.PackedEntity)
	}
	to.Number = uint32(num)

	if bits == 0 {
		return to
	}

	if bits&EntityModel != 0 {
		to.ModelIndex = uint32(m.ReadByte())
	}

	if bits&EntityModel2 != 0 {
		to.ModelIndex2 = uint32(m.ReadByte())
	}

	if bits&EntityModel3 != 0 {
		to.ModelIndex3 = uint32(m.ReadByte())
	}

	if bits&EntityModel4 != 0 {
		to.ModelIndex4 = uint32(m.ReadByte())
	}

	if bits&EntityFrame8 != 0 {
		to.Frame = uint32(m.ReadByte())
	}

	if bits&EntityFrame16 != 0 {
		to.Frame = uint32(m.ReadShort())
	}

	if (bits & (EntitySkin8 | EntitySkin16)) == (EntitySkin8 | EntitySkin16) {
		to.Skin = uint32(m.ReadLong())
	} else if bits&EntitySkin8 != 0 {
		to.Skin = uint32(m.ReadByte())
	} else if bits&EntitySkin16 != 0 {
		to.Skin = uint32(m.ReadWord())
	}

	if (bits & (EntityEffects8 | EntityEffects16)) == (EntityEffects8 | EntityEffects16) {
		to.Effects = uint32(m.ReadLong())
	} else if bits&EntityEffects8 != 0 {
		to.Effects = uint32(m.ReadByte())
	} else if bits&EntityEffects16 != 0 {
		to.Effects = uint32(m.ReadWord())
	}

	if (bits & (EntityRenderFX8 | EntityRenderFX16)) == (EntityRenderFX8 | EntityRenderFX16) {
		to.RenderFx = uint32(m.ReadLong())
	} else if bits&EntityRenderFX8 != 0 {
		to.RenderFx = uint32(m.ReadByte())
	} else if bits&EntityRenderFX16 != 0 {
		to.RenderFx = uint32(m.ReadWord())
	}

	if bits&EntityOrigin1 != 0 {
		to.OriginX = int32(m.ReadShort())
	}

	if bits&EntityOrigin2 != 0 {
		to.OriginY = int32(m.ReadShort())
	}

	if bits&EntityOrigin3 != 0 {
		to.OriginZ = int32(m.ReadShort())
	}

	if bits&EntityAngle1 != 0 {
		to.AngleX = int32(m.ReadByte())
	}

	if bits&EntityAngle2 != 0 {
		to.AngleY = int32(m.ReadByte())
	}

	if bits&EntityAngle3 != 0 {
		to.AngleZ = int32(m.ReadByte())
	}

	if bits&EntityOldOrigin != 0 {
		to.OldOriginX = int32(m.ReadShort())
		to.OldOriginY = int32(m.ReadShort())
		to.OldOriginZ = int32(m.ReadShort())
	}

	if bits&EntitySound != 0 {
		to.Sound = uint32(m.ReadByte())
	}

	if bits&EntityEvent != 0 {
		to.Event = uint32(m.ReadByte())
	}

	if bits&EntitySolid != 0 {
		to.Solid = uint32(m.ReadWord())
	}

	if bits&EntityRemove != 0 {
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