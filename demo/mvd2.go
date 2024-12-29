package demo

import (
	"errors"
	"fmt"
	"os"

	"github.com/packetflinger/libq2/message"
	"github.com/packetflinger/libq2/util"

	pb "github.com/packetflinger/libq2/proto"
)

const (
	MVDMagic         = 843339341 // {'M','V','D','2'}
	MaxMessageLength = 32768     // 0x8000
	MaxPacketLength  = 1400      // default Q2 UDP packet len
	MaxStats         = 32        // playerstate stats uint32 bitmask
	MaxQpath         = 64        // max len for file path
	MaxStringChars   = 1024      // biggest string to deal with
	MaxConfigStrings = 2080      // CS_MAX
	MaxClients       = 256       // hard limit
	MaxEdicts        = 1024      // hard limit
	ProtocolVersion  = 37        // major q2 protocol version
	ProtocolMinimum  = 2009      // minor proto version required
	ProtocolCurrent  = 2010      // minor proto version
	MVD2CMDBits      = 5
	MVD2CMDMask      = ((1 << MVD2CMDBits) - 1)
)

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
	MVDSvcUnicastR
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

// Stream flags (3 bits)
const (
	FlagNoMessages = 1 << 0
	FlagReverved1  = 1 << 1
	FlagReverved2  = 1 << 2
)

type MVD2File struct {
	Filename string
	Handle   *os.File
	Position int // needed?
	Msg      *message.Buffer
}

type MVDGameState struct {
	// node?
	Flags         uint32
	VersionMajor  uint32
	VersionMinor  uint16
	ServerCount   uint32
	ClientNumber  uint16
	GameDir       string
	ConfigStrings []*pb.ConfigString
	// configstring linked list?
	// baseframe linked list?
}

// Open the demo file and return an demo struct.
// Checks if the first 4 bytes match the MVDMagic value.
func OpenMVD2File(f string) (*MVD2File, error) {
	if !util.FileExists(f) {
		return nil, errors.New("no such file")
	}

	fp, e := os.Open(f)
	if e != nil {
		return nil, e
	}

	buffer := make([]byte, 4)
	_, err := fp.Read(buffer)
	if err != nil {
		return nil, err
	}
	msg := message.NewBuffer(buffer)
	if msg.ReadLong() != MVDMagic {
		return nil, errors.New("not a valid multi-view demo file")
	}

	demo := MVD2File{
		Filename: f,
		Handle:   fp,
	}

	return &demo, nil
}

// If the demo file is open (has a valid handle), close it
func (d *MVD2File) Close() {
	if d.Handle != nil {
		d.Handle.Close()
	}
}

/*
func (d *MVD2File) Parse(extcb message.Callback) error {
	//intcb := d.InternalCallbacks()
	for {
		lump, size, err := d.nextLump(d.Handle, int64(d.Position))
		if err != nil {
			return err
		}
		if size == 0 {
			break
		}
		d.Position += size
		buffer := message.NewBuffer(lump)
		d.Msg = &buffer

		d.ParseLump(buffer)
	}
	return nil
}
*/

/*
// Setup the callbacks for demo parsing. Stores data in the appropriate
// spots as it's parsed for later use.
//
// You can't just parse each frame independently, the current frame depends
// on a previous frame for delta compression (usually the last one).
func (d *MVD2File) InternalCallbacks() message.Callback {
	return message.Callback{}
}
*/

func (d *MVD2File) nextLump(f *os.File, pos int64) ([]byte, int, error) {
	_, err := f.Seek(pos, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lumplen := make([]byte, 2)
	_, err = f.Read(lumplen)
	if err != nil {
		return []byte{}, 0, err
	}

	lenbuf := message.Buffer{Data: lumplen, Index: 0}
	length := lenbuf.ReadShort()

	// EOF
	if length == 0 {
		return []byte{}, 0, nil
	}

	_, err = f.Seek(pos+2, 0)
	if err != nil {
		return []byte{}, 0, err
	}

	lump := make([]byte, length)
	read, err := f.Read(lump)
	if err != nil {
		return []byte{}, 0, err
	}

	return lump, read + 2, nil
}

func (d *MVD2File) ParseLump(buf message.Buffer) {
	for buf.Index < len(buf.Data) {
		cmd := buf.ReadByte()
		bits := cmd >> MVD2CMDBits
		cmd &= MVD2CMDMask

		fmt.Println("cmd:", cmd)
		switch cmd {
		case MVDSvcServerData:
			d.ParseServerData(bits)
		case MVDSvcMulticastAll:
			d.ParseMulticast(cmd, uint32(bits))
		}
	}
}

func (d *MVD2File) ParseServerData(bits byte) {
	gs := MVDGameState{}
	gs.VersionMajor = d.Msg.ReadULong() // read unsigned
	gs.VersionMinor = d.Msg.ReadShort()
	gs.ServerCount = d.Msg.ReadULong() // should be unsigned
	gs.GameDir = d.Msg.ReadString()
	gs.ClientNumber = d.Msg.ReadShort()
	gs.Flags = uint32(bits)
	fmt.Println(gs)
	/*
		for {
			cs := d.Msg.ParseConfigString()
			gs.ConfigStrings = append(gs.ConfigStrings, cs)
		}
	*/
}

func (d *MVD2File) ParseMulticast(cmd byte, bits uint32) (*pb.Multicast, error) {
	mc := &pb.Multicast{}
	mc.Type = uint32((cmd - MVDSvcMulticastAll) % 3)
	if cmd < MVDSvcMulticastAllR {
		mc.Reliable = false
	} else {
		mc.Reliable = true
	}
	len := uint32(d.Msg.ReadByte())
	len |= (bits << 8)
	if mc.Type > 0 {
		mc.Leafnum = uint32(d.Msg.ReadShort())
	}
	// parse the actual msg here
	return mc, nil
}

func (d *MVD2File) ParseConfigString() {

}
