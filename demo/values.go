package demo

import (
	pb "github.com/packetflinger/libq2/proto"
)

const (
	MVDMagic          = 843339341 // {'M','V','D','2'}
	GZIPMagic         = 35615     // {0x8b, 0x1f}
	MaxStats          = 32        // playerstate stats uint32 bitmask
	MaxStatsExt       = 64        // extended playerstates supported
	MaxConfigStrings  = 2080      // CS_MAX
	MaxClients        = 256       // hard limit
	MaxEdicts         = 1024      // hard limit
	ProtocolMinimum   = 2009      // minor proto version required
	ProtocolCurrent   = 2010      // minor proto version
	ProtocolPlus      = 2011      // not sure this is used
	ProtocolPlusPlus  = 2012      // same
	ProtocolPlayerFog = 2013
	CommandBits       = 5
	CommandMask       = ((1 << CommandBits) - 1)
	ClientNumNone     = MaxClients - 1
	FlagExtLimits     = 1 << 2 // multiview flags extended limits
	FlagExtLimits2    = 1 << 3 // even more extended!!
)

var (
	// Original (protocol 34/35/36) configstring boundaries
	csRemap = &pb.MvdConfigStringRemap{
		Extended:    false,
		MaxEdicts:   1024,
		MaxModels:   256,
		MaxSounds:   256,
		MaxImages:   256,
		AirAccel:    29,
		MaxClients:  30,
		MapChecksum: 31,
		Models:      32,
		Sounds:      288,
		Images:      544,
		Lights:      800,
		Items:       1056,
		PlayerSkins: 1312,
		General:     1568,
		End:         1568 + (256 * 2),
	}

	// Extended configstring boundaries
	csRemapNew = &pb.MvdConfigStringRemap{
		Extended:    true,
		MaxEdicts:   8192,
		MaxModels:   8192,
		MaxSounds:   2048,
		MaxImages:   2048,
		AirAccel:    59,
		MaxClients:  60,
		MapChecksum: 61,
		Models:      62,
		Sounds:      8254,
		Images:      10302,
		Lights:      12350,
		Items:       12606,
		PlayerSkins: 12862,
		General:     13118,
		End:         13118 + (256 * 2),
	}
)

// Multi-view message types
const (
	MVDSvcBad = iota
	MVDSvcNop
	MVDSvcDisconnect
	MVDSvcReconnect
	MVDSvcServerData
	MVDSvcConfigString
	MVDSvcFrame
	MVDSvcFrameNoDelta
	MVDSvcUnicast
	MVDSvcUnicastReliable
	MVDSvcMulticastAll
	MVDSvcMulticastPHS
	MVDSvcMulticastPVS
	MVDSvcMulticastAllR
	MVDSvcMulticastPHSR
	MVDSvcMulticastPVSR
	MVDSvcSound
	MVDSvcPrint
	MVDSvcStuffText
	MVDSvcMax
)

// Regular demo message types
const (
	SvcBad = iota
	SvcMuzzleFlash
	SvcMuzzleFlash2
	SvcTemporaryEntity
	SvcLayout
	SvcInventory
	SvcNoOperation
	SvcDisconnect
	SvcReconnect
	SvcSound
	SvcPrint
	SvcStuffText
	SvcServerData
	SvcConfigString
	SvcSpawnBaseline
	SvcCenterprint
	SvcDownload
	SvcPlayerInfo
	SvcPacketEntities
	SvcDeltaPacketEntities
	SvcFrame
	SvcMax
)

// Multi-view playerstate bit values
const (
	MvdPlayerType        = 1 << 0
	MvdPlayerOrigin      = 1 << 1
	MvdPlayerOrigin2     = 1 << 2
	MvdPlayerViewOffset  = 1 << 3
	MvdPlayerViewAngles  = 1 << 4
	MvdPlayerViewAngles2 = 1 << 5
	MvdPlayerKickAngles  = 1 << 6
	MvdPlayerBlend       = 1 << 7
	MvdPlayerFov         = 1 << 8
	MvdPlayerWeaponIndex = 1 << 9
	MvdPlayerWeaponFrame = 1 << 10
	MvdPlayerGunOffset   = 1 << 11
	MvdPlayerGunAngles   = 1 << 12
	MvdPlayerRdFlags     = 1 << 13
	MvdPlayerStats       = 1 << 14
	MvdPlayerRemove      = 1 << 15

	MvdPlayerMoreBits = 1 << 8

	MvdPlayerBits = 16
	MvdPlayerMask = (1 << MvdPlayerBits) - 1
)

// Multi-view playerstate flags
const (
	MvdPlayerFlagIgnoreGunIndex    = 1 << 0
	MvdPlayerFlagIgnoreGunFrames   = 1 << 1
	MvdPlayerFlagIgnoreBlend       = 1 << 2
	MvdPlayerFlagIgnoreViewAngles  = 1 << 3
	MvdPlayerFlagIgnoreDeltaAngles = 1 << 4
	MvdPlayerFlagIgnorePrediction  = 1 << 5
	MvdPlayerFlagExtensions        = 1 << 6
	MvdPlayerFlagExtensions2       = 1 << 7
	MvdPlayerFlagForce             = 1 << 8
	MvdPlayerFlagRemove            = 1 << 9
)

// Multi-view stream flags (3 bits)
const (
	FlagNoMessages = 1 << 0
	FlagReverved1  = 1 << 1
	FlagReverved2  = 1 << 2
)

// Multi-view entity bits
const (
	EntityStateLongSolid   = 1 << 3
	EntityStateUMask       = 1 << 4
	EntityStateBeamOrigin  = 1 << 5
	EntityStateShortAngles = 1 << 6
	EntityStateExtensions  = 1 << 7
	EntityStateExtensions2 = 1 << 8

	PlayerStateExtensions  = 1 << 6
	PlayerStateExtensions2 = 1 << 7
)
