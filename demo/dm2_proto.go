package demo

import (
	"github.com/packetflinger/libq2/message"
	pb "github.com/packetflinger/libq2/proto"
)

func (demo *DM2File) Marshal(textpb *pb.TestDemo, data message.MessageBuffer) error {
	//textpb := &pb.TestDemo{}
	currentFrame := &pb.Frame{}
	//fmt.Println("data:\n", data)
	for data.Index < len(data.Buffer) {
		cmd := data.ReadByte()
		switch cmd {
		case message.SVCServerData:
			serverdata := ServerDataToProto(&data)
			//fmt.Println("[svcdata]")
			textpb.Serverinfo = serverdata
		case message.SVCConfigString:
			cs := ConfigstringToProto(&data)
			if currentFrame.GetNumber() == 0 {
				textpb.Configstrings = append(textpb.Configstrings, cs)
				//fmt.Println("[configstring] -", cs.GetIndex())
			} else {
				currentFrame.Configstrings = append(currentFrame.Configstrings, cs)
				//fmt.Println("  [configstring] -", cs.GetIndex())
			}
		case message.SVCSpawnBaseline:
			bitmask := data.ParseEntityBitmask()
			number := data.ParseEntityNumber(bitmask)
			baseline := EntityToProto(&data, bitmask, number)
			//fmt.Println("[baseline] -", number)
			textpb.Baselines = append(textpb.Baselines, baseline)
		case message.SVCStuffText:
			stuff := StuffTextToProto(&data)
			//fmt.Println("[stuff] -", stuff.GetString_())
			if currentFrame.Number > 0 {
				currentFrame.Stufftexts = append(currentFrame.Stufftexts, stuff)
			}
		case message.SVCFrame: // includes playerstate and packetentities
			frame := FrameToProto(&data)
			//fmt.Println(frame)
			//fmt.Println("frameout:", FrameToBinary(frame))
			//fmt.Println("[frame] -", frame.GetNumber())
			textpb.Frames = append(textpb.Frames, frame)
			currentFrame = frame
		case message.SVCPrint:
			print := PrintToProto(&data)
			currentFrame.Prints = append(currentFrame.Prints, print)
			//fmt.Println("[print] -", print.GetString_())
		case message.SVCMuzzleFlash:
			flash := FlashToProto(&data)
			currentFrame.Flashes1 = append(currentFrame.Flashes1, flash)
			//fmt.Println("[muzzleflash]")
		case message.SVCTempEntity:
			te := TempEntToProto(&data)
			currentFrame.TemporaryEntities = append(currentFrame.TemporaryEntities, te)
			//fmt.Println("[temp entity]")
		case message.SVCLayout:
			layout := LayoutToProto(&data)
			currentFrame.Layouts = append(currentFrame.Layouts, layout)
			//fmt.Println("[layout]")
		case message.SVCSound:
			sound := SoundToProto(&data)
			currentFrame.Sounds = append(currentFrame.Sounds, sound)
			//fmt.Println("[sound]")
		case message.SVCCenterPrint:
			cp := CenterPrintToProto(&data)
			currentFrame.Centerprints = append(currentFrame.Centerprints, cp)
			//fmt.Println("[centerprint]")
		}
	}
	return nil
}

func ServerDataToProto(data *message.MessageBuffer) *pb.ServerInfo {
	sd := &pb.ServerInfo{}
	sd.Protocol = data.ReadULong()
	sd.ServerCount = data.ReadULong()
	sd.Demo = data.ReadByte() == 1
	sd.GameDir = data.ReadString()
	sd.ClientNumber = uint32(data.ReadShort())
	sd.MapName = data.ReadString()
	return sd
}

func ServerDataToBinary(s *pb.ServerInfo) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteByte(message.SVCServerData)
	msg.WriteLong(int32(s.GetProtocol()))
	msg.WriteLong(int32(s.GetServerCount()))
	if s.GetDemo() {
		msg.WriteByte(1)
	} else {
		msg.WriteByte(0)
	}
	msg.WriteString(s.GetGameDir())
	msg.WriteShort(uint16(s.GetClientNumber()))
	msg.WriteString(s.GetMapName())
	return msg
}

func ConfigstringToProto(data *message.MessageBuffer) *pb.CString {
	cs := data.ParseConfigString()
	return &pb.CString{
		Index:   uint32(cs.Index),
		String_: cs.String,
	}
}

func ConfigstringToBinary(cs *pb.CString) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteByte(message.SVCConfigString)
	msg.WriteShort(uint16(cs.GetIndex()))
	msg.WriteString(cs.GetString_())
	return msg
}

func EntityToProto(data *message.MessageBuffer, bitmask uint32, number uint16) *pb.PackedEntity {
	b := &pb.PackedEntity{}
	ent := data.ParseEntity(message.PackedEntity{}, number, bitmask)
	b.Number = uint32(number)
	b.ModelIndex = uint32(ent.ModelIndex)
	b.ModelIndex2 = uint32(ent.ModelIndex2)
	b.ModelIndex3 = uint32(ent.ModelIndex3)
	b.ModelIndex4 = uint32(ent.ModelIndex4)
	b.Frame = uint32(ent.Frame)
	b.Skin = ent.SkinNum
	b.Effects = ent.Effects
	b.RenderFx = ent.RenderFX
	b.OriginX = int32(ent.Origin[0])
	b.OriginY = int32(ent.Origin[1])
	b.OriginZ = int32(ent.Origin[2])
	b.AngleX = int32(ent.Angles[0])
	b.AngleY = int32(ent.Angles[1])
	b.AngleZ = int32(ent.Angles[2])
	b.OldOriginX = int32(ent.OldOrigin[0])
	b.OldOriginY = int32(ent.OldOrigin[1])
	b.OldOriginZ = int32(ent.OldOrigin[2])
	b.Sound = uint32(ent.Sound)
	b.Event = uint32(ent.Event)
	b.Solid = ent.Solid
	return b
}

// finish this later
func EntityToBinary(ent *pb.PackedEntity) message.MessageBuffer {
	msg := message.MessageBuffer{}
	return msg
}

func StuffTextToProto(data *message.MessageBuffer) *pb.StuffText {
	st := data.ParseStuffText()
	s := &pb.StuffText{}
	s.String_ = st.String
	return s
}

func StuffTextToBinary(st *pb.StuffText) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteString(st.GetString_())
	return msg
}

func PlayerstateToProto(data *message.MessageBuffer) *pb.PackedPlayer {
	ps := &pb.PackedPlayer{}
	player := data.ParseDeltaPlayerstate(message.PackedPlayer{})
	pm := &pb.PlayerMove{}

	pm.Type = uint32(player.PlayerMove.Type)
	pm.OriginX = int32(player.PlayerMove.Origin[0])
	pm.OriginY = int32(player.PlayerMove.Origin[1])
	pm.OriginZ = int32(player.PlayerMove.Origin[2])
	pm.VelocityX = uint32(player.PlayerMove.Velocity[0])
	pm.VelocityY = uint32(player.PlayerMove.Velocity[1])
	pm.VelocityZ = uint32(player.PlayerMove.Velocity[2])
	pm.Flags = uint32(player.PlayerMove.Flags)
	pm.Time = uint32(player.PlayerMove.Time)
	pm.Gravity = int32(player.PlayerMove.Gravity)
	pm.DeltaAngleX = int32(player.PlayerMove.DeltaAngles[0])
	pm.DeltaAngleY = int32(player.PlayerMove.DeltaAngles[1])
	pm.DeltaAngleZ = int32(player.PlayerMove.DeltaAngles[2])
	ps.Movestate = pm

	ps.ViewAnglesX = int32(player.ViewAngles[0])
	ps.ViewAnglesY = int32(player.ViewAngles[1])
	ps.ViewAnglesZ = int32(player.ViewAngles[2])
	ps.ViewOffsetX = int32(player.ViewOffset[0])
	ps.ViewOffsetY = int32(player.ViewOffset[1])
	ps.ViewOffsetZ = int32(player.ViewOffset[2])
	ps.KickAnglesX = int32(player.KickAngles[0])
	ps.KickAnglesY = int32(player.KickAngles[1])
	ps.KickAnglesZ = int32(player.KickAngles[2])
	ps.GunAnglesX = int32(player.GunAngles[0])
	ps.GunAnglesY = int32(player.GunAngles[1])
	ps.GunAnglesZ = int32(player.GunAngles[2])
	ps.GunOffsetX = int32(player.GunOffset[0])
	ps.GunOffsetY = int32(player.GunOffset[1])
	ps.GunOffsetZ = int32(player.GunOffset[2])
	ps.GunIndex = uint32(player.GunIndex)
	ps.GunFrame = uint32(player.GunFrame)
	ps.BlendW = uint32(player.Blend[0])
	ps.BlendX = uint32(player.Blend[1])
	ps.BlendY = uint32(player.Blend[2])
	ps.BlendZ = uint32(player.Blend[3])
	ps.Fov = uint32(player.FOV)
	ps.RdFlags = uint32(player.RDFlags)

	for i, s := range player.Stats {
		ps.Stats = append(ps.Stats, &pb.PlayerStat{Index: uint32(i), Value: int32(s)})
	}
	return ps
}

// parse a frame first, then playerstate, then delta entities. It should
// always be in that order
func FrameToProto(data *message.MessageBuffer) *pb.Frame {
	frame := data.ParseFrame()
	fr := &pb.Frame{}
	fr.Number = frame.Number
	fr.Delta = frame.Delta
	fr.Suppressed = uint32(frame.Suppressed)
	fr.AreaBytes = uint32(frame.AreaBytes)
	for _, ab := range frame.AreaBits {
		fr.AreaBits = append(fr.AreaBits, uint32(ab))
	}

	ps := &pb.PackedPlayer{}
	if data.ReadByte() == message.SVCPlayerInfo {
		// just delta against a null playerstate
		ps = PlayerstateToProto(data)
	}
	fr.PlayerState = ps

	if data.ReadByte() == message.SVCPacketEntities {
		for {
			bits := data.ParseEntityBitmask()
			num := data.ParseEntityNumber(bits)
			if num <= 0 {
				break
			}
			entity := EntityToProto(data, bits, num)
			fr.Entities = append(fr.Entities, entity)
		}
	}
	return fr
}

func FrameToBinary(fr *pb.Frame) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteByte(message.SVCFrame)
	msg.WriteLong(fr.Number)
	msg.WriteLong(fr.Delta)
	msg.WriteByte(byte(fr.Suppressed))
	msg.WriteByte(byte(fr.AreaBytes))
	for _, ab := range fr.AreaBits {
		msg.WriteByte(byte(ab))
	}

	// from state is empty
	DeltaPlayer(&pb.PackedPlayer{}, fr.PlayerState, &msg)

	msg.WriteByte(message.SVCPacketEntities)
	for _, ent := range fr.GetEntities() {
		DeltaEntity(&pb.PackedEntity{}, ent, &msg)
	}
	msg.WriteShort(0) // EoE
	return msg
}

func DeltaPlayerBitmask(from *pb.PackedPlayer, to *pb.PackedPlayer) uint16 {
	bits := uint16(0)
	mf := from.GetMovestate()
	mt := to.GetMovestate()

	if mf.GetType() != mt.GetType() {
		bits |= message.PlayerType
	}

	if mf.GetOriginX() != mt.GetOriginX() || mf.GetOriginY() != mt.GetOriginY() || mf.GetOriginZ() != mt.GetOriginZ() {
		bits |= message.PlayerOrigin
	}
	if mf.GetVelocityX() != mt.GetVelocityX() || mf.GetVelocityY() != mt.GetVelocityY() || mf.GetVelocityZ() != mt.GetVelocityZ() {
		bits |= message.PlayerVelocity
	}
	if mf.GetTime() != mt.GetTime() {
		bits |= message.PlayerTime
	}
	if mf.GetFlags() != mt.GetFlags() {
		bits |= message.PlayerFlags
	}
	if mf.GetGravity() != mt.GetGravity() {
		bits |= message.PlayerGravity
	}
	if mf.GetDeltaAngleX() != mt.GetDeltaAngleX() || mf.GetDeltaAngleY() != mt.GetDeltaAngleY() || mf.GetDeltaAngleZ() != mt.GetDeltaAngleZ() {
		bits |= message.PlayerDeltaAngles
	}
	if from.GetViewOffsetX() != to.GetViewOffsetX() || from.GetViewOffsetY() != to.GetViewOffsetY() || from.GetViewOffsetZ() != to.GetViewOffsetZ() {
		bits |= message.PlayerViewOffset
	}
	if from.GetViewAnglesX() != to.GetViewAnglesX() || from.GetViewAnglesY() != to.GetViewAnglesY() || from.GetViewAnglesZ() != to.GetViewAnglesZ() {
		bits |= message.PlayerViewAngles
	}
	if from.GetKickAnglesX() != to.GetKickAnglesX() || from.GetKickAnglesY() != to.GetKickAnglesY() || from.GetKickAnglesZ() != to.GetKickAnglesZ() {
		bits |= message.PlayerKickAngles
	}
	if from.GetBlendW() != to.GetBlendW() || from.GetBlendX() != to.GetBlendX() || from.GetBlendY() != to.GetBlendY() || from.GetBlendZ() != to.GetBlendZ() {
		bits |= message.PlayerBlend
	}
	if from.GetFov() != to.GetFov() {
		bits |= message.PlayerFOV
	}
	if from.GetRdFlags() != to.GetRdFlags() {
		bits |= message.PlayerRDFlags
	}
	if from.GetGunFrame() != to.GetGunFrame() || from.GetGunOffsetX() != to.GetGunOffsetX() || from.GetGunOffsetY() != to.GetGunOffsetY() || from.GetGunOffsetZ() != to.GetGunOffsetZ() || from.GetGunAnglesX() != to.GetGunAnglesX() || from.GetGunAnglesY() != to.GetGunAnglesY() || from.GetGunAnglesZ() != to.GetGunAnglesZ() {
		bits |= message.PlayerWeaponFrame
	}
	if from.GetGunIndex() != to.GetGunIndex() {
		bits |= message.PlayerWeaponIndex
	}
	return bits
}

func DeltaPlayer(from *pb.PackedPlayer, to *pb.PackedPlayer, msg *message.MessageBuffer) {
	bits := DeltaPlayerBitmask(from, to)
	msg.WriteByte(message.SVCPlayerInfo)
	msg.WriteShort(bits)

	if bits&message.PlayerType > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetType()))
	}

	if bits&message.PlayerOrigin > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetOriginX()))
		msg.WriteShort(uint16(to.GetMovestate().GetOriginY()))
		msg.WriteShort(uint16(to.GetMovestate().GetOriginZ()))
	}

	if bits&message.PlayerVelocity > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityX()))
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityY()))
		msg.WriteShort(uint16(to.GetMovestate().GetVelocityZ()))
	}

	if bits&message.PlayerTime > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetTime()))
	}

	if bits&message.PlayerFlags > 0 {
		msg.WriteByte(byte(to.GetMovestate().GetFlags()))
	}

	if bits&message.PlayerGravity > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetGravity()))
	}

	if bits&message.PlayerDeltaAngles > 0 {
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleX()))
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleY()))
		msg.WriteShort(uint16(to.GetMovestate().GetDeltaAngleZ()))
	}

	if bits&message.PlayerViewOffset > 0 {
		msg.WriteChar(uint8(to.GetViewOffsetX()))
		msg.WriteChar(uint8(to.GetViewOffsetY()))
		msg.WriteChar(uint8(to.GetViewOffsetZ()))
	}

	if bits&message.PlayerViewAngles > 0 {
		msg.WriteShort(uint16(to.GetViewAnglesX()))
		msg.WriteShort(uint16(to.GetViewAnglesY()))
		msg.WriteShort(uint16(to.GetViewAnglesZ()))
	}

	if bits&message.PlayerKickAngles > 0 {
		msg.WriteChar(uint8(to.GetKickAnglesX()))
		msg.WriteChar(uint8(to.GetKickAnglesY()))
		msg.WriteChar(uint8(to.GetKickAnglesZ()))
	}

	if bits&message.PlayerWeaponIndex > 0 {
		msg.WriteByte(byte(to.GetGunIndex()))
	}

	if bits&message.PlayerWeaponFrame > 0 {
		msg.WriteByte(byte(to.GetGunFrame()))
		msg.WriteChar(uint8(to.GetGunOffsetX()))
		msg.WriteChar(uint8(to.GetGunOffsetY()))
		msg.WriteChar(uint8(to.GetGunOffsetZ()))
		msg.WriteChar(uint8(to.GetGunAnglesX()))
		msg.WriteChar(uint8(to.GetGunAnglesY()))
		msg.WriteChar(uint8(to.GetGunAnglesZ()))
	}

	if bits&message.PlayerFOV > 0 {
		msg.WriteByte(byte(to.GetFov()))
	}

	if bits&message.PlayerRDFlags > 0 {
		msg.WriteByte(byte(to.GetRdFlags()))
	}

	ReconsilePlayerstateStats(to, from, msg)
	/*
		// compress the stats
		statbits := int32(0)
		for i := 0; i < MaxStats; i++ {
			if to.Stats[i] != from.Stats[i] {
				statbits |= 1 << i
			}
		}
	*/
	/*
		stats := to.GetStats()
		msg.WriteLong(statbits)
		for i := 0; i < MaxStats; i++ {
			if (statbits & (1 << i)) > 0 {
				msg.WriteShort(uint16(stats[i]))
			}
		}
	*/
}

func ReconsilePlayerstateStats(to *pb.PackedPlayer, from *pb.PackedPlayer, msg *message.MessageBuffer) {
	bits := uint32(0)
	toStats := [32]int16{}
	for _, s := range to.GetStats() {
		toStats[s.GetIndex()] = int16(s.GetValue())
	}

	fromStats := [32]int16{}
	for _, s := range from.GetStats() {
		fromStats[s.GetIndex()] = int16(s.GetValue())
	}

	for i := 0; i < message.MaxStats; i++ {
		if toStats[i] != fromStats[i] {
			bits |= 1 << i
		}
	}

	msg.WriteLong(int32(bits))

	for i := 0; i < MaxStats; i++ {
		if (bits & (1 << i)) > 0 {
			msg.WriteShort(uint16(toStats[i]))
		}
	}
}

func DeltaEntityBitmask(to *pb.PackedEntity, from *pb.PackedEntity) uint32 {
	bits := uint32(0)
	mask := uint32(0xffff8000)

	if to.GetOriginX() != from.GetOriginX() {
		bits |= message.EntityOrigin1
	}

	if to.GetOriginY() != from.GetOriginY() {
		bits |= message.EntityOrigin2
	}

	if to.GetOriginZ() != from.GetOriginZ() {
		bits |= message.EntityOrigin3
	}

	if to.GetAngleX() != from.GetAngleX() {
		bits |= message.EntityAngle1
	}

	if to.GetAngleY() != from.GetAngleY() {
		bits |= message.EntityAngle2
	}

	if to.GetAngleZ() != from.GetAngleZ() {
		bits |= message.EntityAngle3
	}

	if to.GetSkin() != from.GetSkin() {
		if to.GetSkin()&mask&mask > 0 {
			bits |= message.EntitySkin8 | message.EntitySkin16
		} else if to.GetSkin()&uint32(0x0000ff00) > 0 {
			bits |= message.EntitySkin16
		} else {
			bits |= message.EntitySkin8
		}
	}

	if to.GetFrame() != from.GetFrame() {
		if uint16(to.GetFrame())&uint16(0xff00) > 0 {
			bits |= message.EntityFrame16
		} else {
			bits |= message.EntityFrame8
		}
	}

	if to.Effects != from.Effects {
		if to.Effects&mask > 0 {
			bits |= message.EntityEffects8 | message.EntityEffects16
		} else if to.Effects&0x0000ff00 > 0 {
			bits |= message.EntityEffects16
		} else {
			bits |= message.EntityEffects8
		}
	}

	if to.GetRenderFx() != from.GetRenderFx() {
		if to.GetRenderFx()&mask > 0 {
			bits |= message.EntityRenderFX8 | message.EntityRenderFX16
		} else if to.GetRenderFx()&0x0000ff00 > 0 {
			bits |= message.EntityRenderFX16
		} else {
			bits |= message.EntityRenderFX8
		}
	}

	if to.GetSolid() != from.GetSolid() {
		bits |= message.EntitySolid
	}

	if to.GetEvent() != from.GetEvent() {
		bits |= message.EntityEvent
	}

	if to.GetModelIndex() != from.GetModelIndex() {
		bits |= message.EntityModel
	}

	if to.GetModelIndex2() != from.GetModelIndex2() {
		bits |= message.EntityModel2
	}

	if to.GetModelIndex3() != from.GetModelIndex3() {
		bits |= message.EntityModel3
	}

	if to.GetModelIndex4() != from.GetModelIndex4() {
		bits |= message.EntityModel4
	}

	if to.GetSound() != from.GetSound() {
		bits |= message.EntitySound
	}

	if to.GetRenderFx()&message.RFFrameLerp > 0 {
		bits |= message.EntityOldOrigin
	} else if to.GetRenderFx()&message.RFBeam > 0 {
		bits |= message.EntityOldOrigin
	}

	if to.GetNumber()&0xff00 > 0 {
		bits |= message.EntityNumber16
	}

	if bits&0xff000000 > 0 {
		bits |= message.EntityMoreBits3 | message.EntityMoreBits2 | message.EntityMoreBits1
	} else if bits&0x00ff0000 > 0 {
		bits |= message.EntityMoreBits2 | message.EntityMoreBits1
	} else if bits&0x0000ff00 > 0 {
		bits |= message.EntityMoreBits1
	}

	return bits
}

func DeltaEntity(from *pb.PackedEntity, to *pb.PackedEntity, m *message.MessageBuffer) {
	bits := DeltaEntityBitmask(to, from)

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
	if bits&message.EntityNumber16 > 0 {
		m.WriteShort(uint16(to.GetNumber()))
	} else {
		m.WriteByte(byte(to.GetNumber()))
	}

	if bits&message.EntityModel > 0 {
		m.WriteByte(byte(to.GetModelIndex()))
	}

	if bits&message.EntityModel2 > 0 {
		m.WriteByte(byte(to.GetModelIndex2()))
	}

	if bits&message.EntityModel3 > 0 {
		m.WriteByte(byte(to.GetModelIndex3()))
	}

	if bits&message.EntityModel4 > 0 {
		m.WriteByte(byte(to.GetModelIndex4()))
	}

	if bits&message.EntityFrame8 > 0 {
		m.WriteByte(byte(to.GetFrame()))
	} else if bits&message.EntityFrame16 > 0 {
		m.WriteShort(uint16(to.GetFrame()))
	}

	if (bits & (message.EntitySkin8 | message.EntitySkin16)) == (message.EntitySkin8 | message.EntitySkin16) {
		m.WriteLong(int32(to.GetSkin()))
	} else if bits&message.EntitySkin8 > 0 {
		m.WriteByte(byte(to.GetSkin()))
	} else if bits&message.EntitySkin16 > 0 {
		m.WriteShort(uint16(to.GetSkin()))
	}

	if (bits & (message.EntityEffects8 | message.EntityEffects16)) == (message.EntityEffects8 | message.EntityEffects16) {
		m.WriteLong(int32(to.GetEffects()))
	} else if bits&message.EntityEffects8 > 0 {
		m.WriteByte(byte(to.GetEffects()))
	} else if bits&message.EntityEffects16 > 0 {
		m.WriteShort(uint16(to.GetEffects()))
	}

	if (bits & (message.EntityRenderFX8 | message.EntityRenderFX16)) == (message.EntityRenderFX8 | message.EntityRenderFX16) {
		m.WriteLong(int32(to.GetRenderFx()))
	} else if bits&message.EntityRenderFX8 > 0 {
		m.WriteByte(byte(to.GetRenderFx()))
	} else if bits&message.EntityRenderFX16 > 0 {
		m.WriteShort(uint16(to.GetRenderFx()))
	}

	if bits&message.EntityOrigin1 > 0 {
		m.WriteShort(uint16(to.GetOriginX()))
	}

	if bits&message.EntityOrigin2 > 0 {
		m.WriteShort(uint16(to.GetOriginY()))
	}

	if bits&message.EntityOrigin3 > 0 {
		m.WriteShort(uint16(to.GetOriginZ()))
	}

	if bits&message.EntityAngle1 > 0 {
		m.WriteByte(byte(to.GetAngleX() >> 8))
	}

	if bits&message.EntityAngle2 > 0 {
		m.WriteByte(byte(to.GetAngleY() >> 8))
	}

	if bits&message.EntityAngle3 > 0 {
		m.WriteByte(byte(to.GetAngleZ() >> 8))
	}

	if bits&message.EntityOldOrigin > 0 {
		m.WriteShort(uint16(to.GetOldOriginX()))
		m.WriteShort(uint16(to.GetOldOriginY()))
		m.WriteShort(uint16(to.GetOldOriginZ()))
	}

	if bits&message.EntitySound > 0 {
		m.WriteByte(byte(to.GetSound()))
	}

	if bits&message.EntityEvent > 0 {
		m.WriteByte(byte(to.GetEvent()))
	}

	if bits&message.EntitySolid > 0 {
		m.WriteShort(uint16(to.GetSolid()))
	}
}

func PrintToProto(data *message.MessageBuffer) *pb.Print {
	pr := data.ParsePrint()
	p := &pb.Print{}
	p.Level = uint32(pr.Level)
	p.String_ = pr.String
	return p
}

func PrintToBinary(p *pb.Print) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteShort(uint16(p.GetLevel()))
	msg.WriteString(p.GetString_())
	return msg
}

func FlashToProto(data *message.MessageBuffer) *pb.MuzzleFlash {
	f := &pb.MuzzleFlash{}
	mf := data.ParseMuzzleFlash()
	f.Entity = uint32(mf.Entity)
	f.Weapon = uint32(mf.Weapon)
	return f
}

func FlashToBinary(mf *pb.MuzzleFlash) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteShort(uint16(mf.GetEntity()))
	msg.WriteByte(byte(mf.GetWeapon()))
	return msg
}

func TempEntToProto(data *message.MessageBuffer) *pb.TemporaryEntity {
	t := &pb.TemporaryEntity{}
	te := data.ParseTempEntity()
	t.Type = uint32(te.Type)
	t.Position1X = uint32(te.Position1[0])
	t.Position1Y = uint32(te.Position1[1])
	t.Position1Z = uint32(te.Position1[2])
	t.Position2X = uint32(te.Position2[0])
	t.Position2Y = uint32(te.Position2[1])
	t.Position2Z = uint32(te.Position2[2])
	t.OffsetX = uint32(te.Offset[0])
	t.OffsetY = uint32(te.Offset[1])
	t.OffsetZ = uint32(te.Offset[2])
	t.Direction = uint32(te.Direction)
	t.Count = uint32(te.Count)
	t.Color = uint32(te.Color)
	t.Entity1 = uint32(te.Entity1)
	t.Entity2 = uint32(te.Entity2)
	t.Time = uint32(te.Time)
	return t
}

func TempEntToBinary(te *pb.TemporaryEntity) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteByte(byte(te.GetType()))
	switch te.GetType() {
	case message.TentBlood:
		fallthrough
	case message.TentGunshot:
		fallthrough
	case message.TentSparks:
		fallthrough
	case message.TentBulletSparks:
		fallthrough
	case message.TentScreenSparks:
		fallthrough
	case message.TentShieldSparks:
		fallthrough
	case message.TentShotgun:
		fallthrough
	case message.TentBlaster:
		fallthrough
	case message.TentGreenBlood:
		fallthrough
	case message.TentBlaster2:
		fallthrough
	case message.TentFlechette:
		fallthrough
	case message.TentHeatBeamSparks:
		fallthrough
	case message.TentHeatBeamSteam:
		fallthrough
	case message.TentMoreBlood:
		fallthrough
	case message.TentElectricSparks:
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteByte(byte(te.GetDirection()))
	case message.TentSplash:
		fallthrough
	case message.TentLaserSparks:
		fallthrough
	case message.TentWeldingSparks:
		fallthrough
	case message.TentTunnelSparks:
		msg.WriteByte(byte(te.GetCount()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteByte(byte(te.GetDirection()))
		msg.WriteByte(byte(te.GetColor()))
	case message.TentBlueHyperBlaster:
		fallthrough
	case message.TentRailTrail:
		fallthrough
	case message.TentBubbleTrail:
		fallthrough
	case message.TentDebugTrail:
		fallthrough
	case message.TentBubbleTrail2:
		fallthrough
	case message.TentBFGLaser:
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteCoord(uint16(te.GetPosition2X()))
		msg.WriteCoord(uint16(te.GetPosition2Y()))
		msg.WriteCoord(uint16(te.GetPosition2Z()))
	case message.TentGrenadeExplosion:
		fallthrough
	case message.TentGrenadeExplosionWater:
		fallthrough
	case message.TentExplosion2:
		fallthrough
	case message.TentPlasmaExplosion:
		fallthrough
	case message.TentRocketExplosion:
		fallthrough
	case message.TentRocketExplosionWater:
		fallthrough
	case message.TentExplosion1:
		fallthrough
	case message.TentExplosion1NP:
		fallthrough
	case message.TentExplosion1Big:
		fallthrough
	case message.TentBFGExplosion:
		fallthrough
	case message.TentBFGBigExplosion:
		fallthrough
	case message.TentBossTeleport:
		fallthrough
	case message.TentPlainExplosion:
		fallthrough
	case message.TentChainFistSmoke:
		fallthrough
	case message.TentTrackerExplosion:
		fallthrough
	case message.TentTeleportEffect:
		fallthrough
	case message.TentDBallGoal:
		fallthrough
	case message.TentWidowSplash:
		fallthrough
	case message.TentNukeBlast:
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
	case message.TentParasiteAttack:
		fallthrough
	case message.TentMedicCableAttack:
		fallthrough
	case message.TentHeatBeam:
		fallthrough
	case message.TentMonsterHeatBeam:
		msg.WriteShort(uint16(te.GetEntity1()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteCoord(uint16(te.GetPosition2X()))
		msg.WriteCoord(uint16(te.GetPosition2Y()))
		msg.WriteCoord(uint16(te.GetPosition2Z()))
		msg.WriteCoord(uint16(te.GetOffsetX()))
		msg.WriteCoord(uint16(te.GetOffsetY()))
		msg.WriteCoord(uint16(te.GetOffsetZ()))
	case message.TentGrappleCable:
		msg.WriteShort(uint16(te.GetEntity1()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteCoord(uint16(te.GetPosition2X()))
		msg.WriteCoord(uint16(te.GetPosition2Y()))
		msg.WriteCoord(uint16(te.GetPosition2Z()))
		msg.WriteCoord(uint16(te.GetOffsetX()))
		msg.WriteCoord(uint16(te.GetOffsetY()))
		msg.WriteCoord(uint16(te.GetOffsetZ()))
	case message.TentLightning:
		msg.WriteShort(uint16(te.GetEntity1()))
		msg.WriteShort(uint16(te.GetEntity2()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteCoord(uint16(te.GetPosition2X()))
		msg.WriteCoord(uint16(te.GetPosition2Y()))
		msg.WriteCoord(uint16(te.GetPosition2Z()))
	case message.TentFlashlight:
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteShort(uint16(te.GetEntity1()))
	case message.TentForceWall:
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteCoord(uint16(te.GetPosition2X()))
		msg.WriteCoord(uint16(te.GetPosition2Y()))
		msg.WriteCoord(uint16(te.GetPosition2Z()))
		msg.WriteByte(byte(te.GetColor()))
	case message.TentSteam:
		msg.WriteShort(uint16(te.GetEntity1()))
		msg.WriteByte(byte(te.GetCount()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
		msg.WriteByte(byte(te.GetDirection()))
		msg.WriteByte(byte(te.GetColor()))
		msg.WriteShort(uint16(te.GetEntity2()))
		if int32(te.Entity1) != -1 {
			msg.WriteLong(int32(te.GetTime()))
		}
	case message.TentWidowBeamOut:
		msg.WriteShort(uint16(te.GetEntity1()))
		msg.WriteCoord(uint16(te.GetPosition1X()))
		msg.WriteCoord(uint16(te.GetPosition1Y()))
		msg.WriteCoord(uint16(te.GetPosition1Z()))
	}
	return msg
}

func LayoutToProto(data *message.MessageBuffer) *pb.Layout {
	lo := &pb.Layout{}
	layout := data.ParseLayout()
	lo.String_ = layout.Data
	return lo
}

func LayoutToBinary(lo *pb.Layout) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteString(lo.GetString_())
	return msg
}

func SoundToProto(data *message.MessageBuffer) *pb.PackedSound {
	sound := &pb.PackedSound{}
	s := data.ParseSound()
	sound.Flags = uint32(s.Flags)
	sound.Index = uint32(s.Index)
	sound.Volume = uint32(s.Volume)
	sound.Attenuation = uint32(s.Attenuation)
	sound.TimeOffset = uint32(s.TimeOffset)
	sound.Channel = uint32(s.Channel)
	sound.Entity = uint32(s.Entity)
	sound.PositionX = uint32(s.Position[0])
	sound.PositionY = uint32(s.Position[1])
	sound.PositionZ = uint32(s.Position[2])
	return sound
}

func SoundToBinary(s *pb.PackedSound) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteByte(byte(s.GetFlags()))
	msg.WriteByte(byte(s.GetIndex()))
	if (s.GetFlags() & message.SoundVolume) > 0 {
		msg.WriteByte(byte(s.GetVolume()))
	}

	if (s.GetFlags() & message.SoundAttenuation) > 0 {
		msg.WriteByte(byte(s.GetAttenuation()))
	}

	if (s.GetFlags() & message.SoundOffset) > 0 {
		msg.WriteByte(byte(s.GetTimeOffset()))
	}

	// fix this
	if (s.GetFlags() & message.SoundEntity) > 0 {
		//s.Channel = m.ReadShort() & 7
		//s.Entity = s.Channel >> 3
		msg.WriteShort(uint16(s.GetEntity()<<3 + s.GetChannel()))
	}

	if (s.GetFlags() & message.SoundPosition) > 0 {
		msg.WriteCoord(uint16(s.GetPositionX()))
		msg.WriteCoord(uint16(s.GetPositionY()))
		msg.WriteCoord(uint16(s.GetPositionZ()))
	}
	return msg
}

func CenterPrintToProto(data *message.MessageBuffer) *pb.CenterPrint {
	cp := &pb.CenterPrint{}
	center := data.ParseCenterPrint()
	cp.String_ = center.Data
	return cp
}

func CenterPrintToBinary(cp *pb.CenterPrint) message.MessageBuffer {
	msg := message.MessageBuffer{}
	msg.WriteString(cp.GetString_())
	return msg
}
