package demo

import (
	"os"
	"slices"

	"github.com/packetflinger/libq2/message"
	pb "github.com/packetflinger/libq2/proto"
)

// MVD2Writer builds a binary .mvd2 stream from textproto-based MvdDemo data.
// It tracks the same kind of running state the reader does (last known player
// and entity states) so that successive calls emit correct deltas.
type MVD2Writer struct {
	demo *pb.MvdDemo
}

// Creates a new writer struct. All the proto-to-binary writing funcs use this
// struct as a receiver.
func NewMVD2Writer(mvd *pb.MvdDemo) *MVD2Writer {
	if mvd == nil {
		mvd = &pb.MvdDemo{}
	}
	if mvd.Players == nil {
		mvd.Players = make(map[int32]*pb.MvdPlayer)
	}
	if mvd.Entities == nil {
		mvd.Entities = make(map[int32]*pb.PackedEntity)
	}
	return &MVD2Writer{
		demo: mvd,
	}
}

// This is the final step when writing a demo. It will write all the binary
// data generated to a file named from the argument.
func (w *MVD2Writer) Finalize(name string) error {
	out, err := w.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(name, out.Data, 0644)
}

// Marshal is the top level function for converting a demo (as parsed by
// MVD2Parser.Unmarshal, or built up by hand) back into the binary .mvd2
// format: the magic header, one length-prefixed packet per pb.MvdPacket, and
// the trailing zero-length end-of-file marker.
//
// Because MvdPacket buckets messages by type rather than recording the exact
// interleaving they originally arrived in, a re-marshaled file isn't
// guaranteed to be byte-identical to whatever produced the parsed demo -- but
// it is a valid, symmetric .mvd2 stream that MVD2Parser can read back.
func (w *MVD2Writer) Marshal() (message.Buffer, error) {
	out := message.NewBuffer(nil)
	out.WriteLong(MVDMagic)
	for _, pkt := range w.demo.GetPackets() {
		body := w.MarshalPacket(pkt)
		out.WriteShort(body.Size())
		out.Append(body)
	}
	out.WriteShort(0) // end of file
	return out, nil
}

// MarshalPacket assembles one top-level (length-prefixed) packet's worth of
// messages: serverdata+configstrings+baseline frame (if this packet starts a
// new demo/map), any further frames, and any sounds/prints/unicasts/
// multicasts collected for it.
func (w *MVD2Writer) MarshalPacket(pkt *pb.MvdPacket) message.Buffer {
	out := message.NewBuffer(nil)
	frames := pkt.GetFrames()

	if sd := pkt.GetServerdata(); sd != nil {
		// A serverdata packet starts a new demo/map: sync the writer's
		// protocol state from it, mirroring how ParseServerData updates
		// p.demo, so configstrings/players/entities encode correctly.
		w.demo.Remap = sd.GetRemap()
		w.demo.EntityStateFlags = sd.GetEntitystateFlags()
		w.demo.PlayerStateFlags = sd.GetPlayerstateFlags()
		w.demo.Players = make(map[int32]*pb.MvdPlayer)
		w.demo.Entities = make(map[int32]*pb.PackedEntity)

		out.Append(w.MarshalServerData(sd))
		out.Append(w.MarshalConfigstrings(pkt.GetConfigstrings()))
		if len(frames) > 0 {
			// The baseline frame that immediately follows a serverdata isn't
			// prefixed with its own MVDSvcFrame command byte.
			out.Append(w.MarshalFrame(frames[0]))
			frames = frames[1:]
		}
	} else {
		for _, cs := range pkt.GetConfigstrings() {
			out.Append(w.MarshalConfigStringUpdate(cs))
		}
	}

	for _, frame := range frames {
		out.WriteByte(MVDSvcFrame)
		out.Append(w.MarshalFrame(frame))
	}
	for _, snd := range pkt.GetSounds() {
		out.Append(w.MarshalSound(snd))
	}
	for _, pr := range pkt.GetPrints() {
		out.Append(w.MarshalPrint(pr))
	}
	for _, uc := range pkt.GetUnicasts() {
		out.Append(w.MarshalUnicast(uc, false))
	}
	for _, mc := range pkt.GetMulticasts() {
		buf, _ := w.MarshalMulticast(mc)
		out.Append(*buf)
	}
	return out
}

// MarshalServerData writes an MvdServerData message, embedding it in the
// same MVDSvcServerData command byte used everywhere else. Pre-PlusPlus
// protocols carry their flags in the command byte's extra bits instead of an
// explicit field, exactly like ParseServerData expects.
func (w *MVD2Writer) MarshalServerData(sd *pb.MvdServerData) message.Buffer {
	out := message.NewBuffer(nil)
	cmd := MVDSvcServerData
	if sd.GetProtocol() < ProtocolPlusPlus {
		cmd |= int(sd.GetFlags()&CommandMask) << CommandBits
	}
	out.WriteByte(cmd)
	out.WriteLongP(37)
	out.WriteShortP(sd.GetProtocol())
	if sd.GetProtocol() >= ProtocolPlusPlus {
		out.WriteShortP(sd.GetFlags())
	}
	out.WriteLongP(sd.GetIdentity())
	out.WriteString(sd.GetGameDirectory())
	out.WriteShortP(sd.GetDummyClient())
	return out
}

// MarshalConfigstrings writes the blob of configstrings that follows
// serverdata at the start of a demo, terminated with the remap's end marker.
func (w *MVD2Writer) MarshalConfigstrings(data map[int32]*pb.ConfigString) message.Buffer {
	out := message.NewBuffer(nil)
	var keys []int32
	for k := range data {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		out.Append(w.MarshalConfigstring(data[k]))
	}
	out.WriteShortP(w.demo.GetRemap().GetEnd())
	return out
}

// MarshalConfigstring writes a single index+string pair, as found in the
// gamestate's configstring blob.
func (w *MVD2Writer) MarshalConfigstring(data *pb.ConfigString) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteShortP(int32(data.GetIndex()))
	out.WriteString(data.GetData())
	return out
}

// MarshalConfigStringUpdate writes a standalone MVDSvcConfigString command,
// used for a single configstring change outside of the initial gamestate.
func (w *MVD2Writer) MarshalConfigStringUpdate(cs *pb.ConfigString) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteByte(MVDSvcConfigString)
	out.Append(w.MarshalConfigstring(cs))
	return out
}

// MarshalPrint writes a standalone MVDSvcPrint command.
func (w *MVD2Writer) MarshalPrint(pr *pb.Print) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteByte(MVDSvcPrint)
	out.WriteByteP(pr.GetLevel())
	out.WriteString(pr.GetData())
	return out
}

// MarshalSound generates a full MVDSvcSound command from a PackedSound proto,
// symmetric with MVD2Parser.ParseSound.
func (w *MVD2Writer) MarshalSound(sound *pb.PackedSound) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteByte(MVDSvcSound)

	flags := sound.GetFlags()
	if sound.GetIndex() > 255 {
		flags |= message.SoundIndex16
	}
	out.WriteByteP(flags)
	if sound.GetIndex() > 255 {
		out.WriteWordP(sound.GetIndex())
	} else {
		out.WriteByteP(sound.GetIndex())
	}
	if (flags & message.SoundVolume) != 0 {
		out.WriteByteP(sound.GetVolume())
	}
	if (flags & message.SoundAttenuation) != 0 {
		out.WriteByteP(sound.GetAttenuation())
	}
	if (flags & message.SoundOffset) != 0 {
		out.WriteByteP(sound.GetTimeOffset())
	}
	sendchan := (sound.GetEntity() << 3) | (sound.GetChannel() & 0x7)
	out.WriteWordP(sendchan)
	return out
}

// MarshalMulticast generates a Multicast proto back into its MVDSvcMulticast*
// command, symmetric with MVD2Parser.ParseMulticast. mc.Type must hold one of
// the raw MVDSvcMulticast* command values (as ParseMulticast fills in).
func (w *MVD2Writer) MarshalMulticast(mc *pb.MvdMulticast) (*message.Buffer, error) {
	out := message.NewBuffer(nil)
	cmd := int(mc.GetType())
	length := len(mc.GetData())
	extra := (length >> 8) & CommandMask
	out.WriteByte(cmd | (extra << CommandBits))
	out.WriteByte(length & 0xff)
	switch cmd {
	case MVDSvcMulticastPHS, MVDSvcMulticastPVS, MVDSvcMulticastPHSR, MVDSvcMulticastPVSR:
		out.WriteWordP(uint32(mc.GetLeaf()))
	}
	out.WriteData(mc.GetData())
	return &out, nil
}

// MarshalUnicast generates an MVDSvcUnicast(Reliable) command from an
// MvdUnicast proto. Since MvdUnicast buckets its sub-messages by type instead
// of recording their original interleaving, the emitted order is always
// layouts, then configstrings, then prints, then stufftexts.
func (w *MVD2Writer) MarshalUnicast(u *pb.MvdUnicast, reliable bool) message.Buffer {
	payload := message.NewBuffer(nil)
	for _, lo := range u.GetLayouts() {
		payload.WriteByte(SvcLayout)
		payload.WriteString(lo.GetData())
	}
	for _, cs := range u.GetConfigstrings() {
		payload.WriteByte(SvcConfigString)
		payload.WriteWordP(cs.GetIndex())
		payload.WriteString(cs.GetData())
	}
	for _, pr := range u.GetPrints() {
		payload.WriteByte(SvcPrint)
		payload.WriteByteP(pr.GetLevel())
		payload.WriteString(pr.GetData())
	}
	for _, st := range u.GetStuffs() {
		payload.WriteByte(SvcStuffText)
		payload.WriteString(st.GetData())
	}

	cmd := MVDSvcUnicast
	if reliable {
		cmd = MVDSvcUnicastReliable
	}
	length := payload.Size()
	extra := (length >> 8) & CommandMask

	out := message.NewBuffer(nil)
	out.WriteByte(cmd | (extra << CommandBits))
	out.WriteByte(length & 0xff)
	out.WriteByte(int(u.GetClientNumber()))
	out.Append(payload)
	return out
}

// MarshalFrame writes one frame: portal bits, then all packet players, then
// all packet entities.
func (w *MVD2Writer) MarshalFrame(frame *pb.MvdFrame) message.Buffer {
	out := message.NewBuffer(nil)
	out.WriteByte(len(frame.GetPortalData()))
	out.WriteData(frame.GetPortalData())
	out.Append(w.MarshalPlayers(frame.GetPlayers()))
	out.Append(w.MarshalEntities(frame.GetEntities()))
	return out
}

// MarshalPlayers writes every player present this frame, terminated by the
// CLIENTNUM_NONE sentinel byte.
func (w *MVD2Writer) MarshalPlayers(players map[int32]*pb.PackedPlayer) message.Buffer {
	out := message.NewBuffer(nil)
	var keys []int32
	for k := range players {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, num := range keys {
		out.Append(w.MarshalPlayer(num, players[num]))
	}
	out.WriteByte(ClientNumNone)
	return out
}

// MarshalPlayer writes one player's delta record against the writer's last
// known state for that player number, then updates that state so the next
// call deltas correctly. Passing a nil `to` marks the player removed.
func (w *MVD2Writer) MarshalPlayer(num int32, to *pb.PackedPlayer) message.Buffer {
	existing := w.demo.GetPlayers()[num]
	from := existing.GetPlayerState()
	flags := w.demo.GetPlayerStateFlags()

	out := writeDeltaPlayer(num, from, to, flags)

	pl := existing
	if pl == nil {
		pl = &pb.MvdPlayer{}
		w.demo.Players[num] = pl
	}
	if to == nil {
		pl.InUse = false
	} else {
		pl.PlayerState = to
		pl.InUse = true
	}
	return out
}

// MarshalEntities writes every entity touched this frame, in ascending
// number order, terminated by a zero bitmask/number word. Entities are
// delta-encoded against the writer's running baseline, which is updated as
// each one is written -- mirroring MVD2Parser.ParseDeltaEntities.
//
// Extended entity protocols (long solid, short angles, and the extensions/
// extensions2 fields) aren't supported by the writer yet; it always emits the
// base (protocol < Plus) wire format.
func (w *MVD2Writer) MarshalEntities(ents map[int32]*pb.PackedEntity) message.Buffer {
	out := message.NewBuffer(nil)

	// ents need to be in numeric order and maps are not guaranteed to give
	// their values in the order they were added. So export the keys and
	// sort them.
	var keys []int32
	for k := range ents {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		to := ents[k]
		from := w.demo.GetEntities()[k]
		out.Append(message.WriteDeltaEntity(from, to))
		// Keep the entry as the delta baseline even when removed, matching
		// ParseDeltaEntities: q2pro doesn't clear the underlying
		// entity_state_t on removal either.
		w.demo.Entities[k] = to
	}
	out.WriteShort(0) // combined bitmask and number
	return out
}

// deltaPlayerBitmask computes the PPS_* bitmask (see values.go) of fields
// that differ between `from` and `to`, mirroring ParseDeltaPlayer's read
// side. `from` may be nil, in which case every field `to` sets is considered
// changed -- matching how a freshly seen player decodes from a nil baseline.
func deltaPlayerBitmask(from, to *pb.PackedPlayer, flags int32) uint32 {
	var bits uint32
	fm, tm := from.GetMovestate(), to.GetMovestate()

	if fm.GetType() != tm.GetType() {
		bits |= MvdPlayerType
	}
	if fm.GetOriginX() != tm.GetOriginX() || fm.GetOriginY() != tm.GetOriginY() {
		bits |= MvdPlayerOrigin
	}
	if fm.GetOriginZ() != tm.GetOriginZ() {
		bits |= MvdPlayerOrigin2
	}
	if from.GetViewOffsetX() != to.GetViewOffsetX() ||
		from.GetViewOffsetY() != to.GetViewOffsetY() ||
		from.GetViewOffsetZ() != to.GetViewOffsetZ() {
		bits |= MvdPlayerViewOffset
	}
	if from.GetViewAnglesX() != to.GetViewAnglesX() || from.GetViewAnglesY() != to.GetViewAnglesY() {
		bits |= MvdPlayerViewAngles
	}
	if from.GetViewAnglesZ() != to.GetViewAnglesZ() {
		bits |= MvdPlayerViewAngles2
	}
	if from.GetKickAnglesX() != to.GetKickAnglesX() ||
		from.GetKickAnglesY() != to.GetKickAnglesY() ||
		from.GetKickAnglesZ() != to.GetKickAnglesZ() {
		bits |= MvdPlayerKickAngles
	}
	if from.GetGunIndex() != to.GetGunIndex() {
		bits |= MvdPlayerWeaponIndex
	}
	if from.GetGunFrame() != to.GetGunFrame() {
		bits |= MvdPlayerWeaponFrame
	}
	if from.GetGunOffsetX() != to.GetGunOffsetX() ||
		from.GetGunOffsetY() != to.GetGunOffsetY() ||
		from.GetGunOffsetZ() != to.GetGunOffsetZ() {
		bits |= MvdPlayerGunOffset
	}
	if from.GetGunAnglesX() != to.GetGunAnglesX() ||
		from.GetGunAnglesY() != to.GetGunAnglesY() ||
		from.GetGunAnglesZ() != to.GetGunAnglesZ() {
		bits |= MvdPlayerGunAngles
	}
	if from.GetBlendW() != to.GetBlendW() || from.GetBlendX() != to.GetBlendX() ||
		from.GetBlendY() != to.GetBlendY() || from.GetBlendZ() != to.GetBlendZ() ||
		from.GetDamageBlendW() != to.GetDamageBlendW() || from.GetDamageBlendX() != to.GetDamageBlendX() ||
		from.GetDamageBlendY() != to.GetDamageBlendY() || from.GetDamageBlendZ() != to.GetDamageBlendZ() {
		bits |= MvdPlayerBlend
	}
	if from.GetFov() != to.GetFov() {
		bits |= MvdPlayerFov
	}
	if from.GetRdFlags() != to.GetRdFlags() {
		bits |= MvdPlayerRdFlags
	}
	if statsDiffer(from.GetStats(), to.GetStats(), flags) {
		bits |= MvdPlayerStats
	}
	return bits
}

func statsDiffer(from, to map[uint32]int32, flags int32) bool {
	num := uint32(MaxStats)
	if flags&MvdPlayerFlagExtensions2 != 0 {
		num = MaxStatsExt
	}
	for i := uint32(0); i < num; i++ {
		if from[i] != to[i] {
			return true
		}
	}
	return false
}

// writeDeltaPlayer emits one player's delta record (client number, bits,
// changed fields), mirroring ParseDeltaPlayer/ParsePacketPlayers' read side.
// A nil `to` writes just the "player removed" marker, matching
// MSG_WriteDeltaPlayerstate_Packet's behavior for a dropped player.
func writeDeltaPlayer(num int32, from, to *pb.PackedPlayer, flags int32) message.Buffer {
	out := message.NewBuffer(nil)
	moreBits := flags&MvdPlayerFlagMoreBits != 0

	if to == nil {
		out.WriteByte(int(num))
		out.WriteWord(MvdPlayerMoreBits) // MOREBITS doubles as REMOVE for old demos
		if moreBits {
			out.WriteByte(MvdPlayerRemove >> 16)
		}
		return out
	}

	bits := deltaPlayerBitmask(from, to, flags)
	if moreBits && bits > 0x7fff {
		bits |= MvdPlayerMoreBits
	}

	out.WriteByte(int(num))
	out.WriteWord(int(bits & 0xffff))
	if moreBits && (bits&MvdPlayerMoreBits) != 0 {
		out.WriteByte(int((bits >> 16) & 0xff))
	}

	tm := to.GetMovestate()
	ext2 := flags&MvdPlayerFlagExtensions2 != 0
	if (bits & MvdPlayerType) != 0 {
		out.WriteByteP(tm.GetType())
	}
	if ext2 {
		if (bits & MvdPlayerOrigin) != 0 {
			writeExtCoord(&out, tm.GetOriginX())
			writeExtCoord(&out, tm.GetOriginY())
		}
		if (bits & MvdPlayerOrigin2) != 0 {
			writeExtCoord(&out, tm.GetOriginZ())
		}
	} else {
		if (bits & MvdPlayerOrigin) != 0 {
			out.WriteShortP(tm.GetOriginX())
			out.WriteShortP(tm.GetOriginY())
		}
		if (bits & MvdPlayerOrigin2) != 0 {
			out.WriteShortP(tm.GetOriginZ())
		}
	}
	if (bits & MvdPlayerViewOffset) != 0 {
		out.WriteCharP(to.GetViewOffsetX())
		out.WriteCharP(to.GetViewOffsetY())
		out.WriteCharP(to.GetViewOffsetZ())
	}
	if (bits & MvdPlayerViewAngles) != 0 {
		out.WriteShortP(to.GetViewAnglesX())
		out.WriteShortP(to.GetViewAnglesY())
	}
	if (bits & MvdPlayerViewAngles2) != 0 {
		out.WriteShortP(to.GetViewAnglesZ())
	}
	if (bits & MvdPlayerKickAngles) != 0 {
		out.WriteCharP(to.GetKickAnglesX())
		out.WriteCharP(to.GetKickAnglesY())
		out.WriteCharP(to.GetKickAnglesZ())
	}
	if (bits & MvdPlayerWeaponIndex) != 0 {
		if flags&MvdPlayerFlagExtensions != 0 {
			out.WriteWordP(to.GetGunIndex())
		} else {
			out.WriteByteP(to.GetGunIndex())
		}
	}
	if (bits & MvdPlayerWeaponFrame) != 0 {
		out.WriteByteP(to.GetGunFrame())
	}
	if (bits & MvdPlayerGunOffset) != 0 {
		out.WriteCharP(to.GetGunOffsetX())
		out.WriteCharP(to.GetGunOffsetY())
		out.WriteCharP(to.GetGunOffsetZ())
	}
	if (bits & MvdPlayerGunAngles) != 0 {
		out.WriteCharP(to.GetGunAnglesX())
		out.WriteCharP(to.GetGunAnglesY())
		out.WriteCharP(to.GetGunAnglesZ())
	}
	if (bits & MvdPlayerBlend) != 0 {
		if ext2 {
			out.WriteByte(0xff) // send every blend/damage_blend channel
			out.WriteByte(int(to.GetBlendW()))
			out.WriteByte(int(to.GetBlendX()))
			out.WriteByte(int(to.GetBlendY()))
			out.WriteByte(int(to.GetBlendZ()))
			out.WriteByte(int(to.GetDamageBlendW()))
			out.WriteByte(int(to.GetDamageBlendX()))
			out.WriteByte(int(to.GetDamageBlendY()))
			out.WriteByte(int(to.GetDamageBlendZ()))
		} else {
			out.WriteByte(int(to.GetBlendW()))
			out.WriteByte(int(to.GetBlendX()))
			out.WriteByte(int(to.GetBlendY()))
			out.WriteByte(int(to.GetBlendZ()))
		}
	}
	if (bits & MvdPlayerFov) != 0 {
		out.WriteByteP(to.GetFov())
	}
	if (bits & MvdPlayerRdFlags) != 0 {
		out.WriteByteP(to.GetRdFlags())
	}
	if (bits & MvdPlayerStats) != 0 {
		writePlayerStats(&out, from.GetStats(), to.GetStats(), flags)
	}
	return out
}

// writePlayerStats writes the changed-stats bitmask (32 or 64 bits depending
// on extensions2) followed by each changed stat's new value.
func writePlayerStats(out *message.Buffer, from, to map[uint32]int32, flags int32) {
	ext2 := flags&MvdPlayerFlagExtensions2 != 0
	num := uint32(MaxStats)
	if ext2 {
		num = MaxStatsExt
	}
	var bits uint64
	for i := uint32(0); i < num; i++ {
		if from[i] != to[i] {
			bits |= 1 << i
		}
	}
	if ext2 {
		writeVarInt64(out, bits)
	} else {
		out.WriteLong(int(int32(bits)))
	}
	for i := uint32(0); i < num; i++ {
		if (bits & (1 << i)) != 0 {
			out.WriteShortP(to[i])
		}
	}
}

// writeVarInt64 encodes v as a 7-bits-per-byte varint with a continuation bit,
// matching Buffer.ReadVarInt64.
func writeVarInt64(out *message.Buffer, v uint64) {
	for {
		b := int(v & 0x7f)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		out.WriteByte(b)
		if v == 0 {
			return
		}
	}
}

// writeExtCoord writes an extensions2 (protocol++) coordinate in the
// "absolute value" wire form -- always valid and always decodable regardless
// of the reader's delta base, just not maximally compact. Mirrors the long
// form of q2pro's MSG_WriteDeltaInt23/MSG_WriteIntPos.
func writeExtCoord(out *message.Buffer, value int32) {
	v := (uint32(value) << 1) | 1
	out.WriteWord(int(v & 0xffff))
	out.WriteByte(int((v >> 16) & 0xff))
}
