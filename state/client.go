package state

import (
	"github.com/packetflinger/libq2/client"
)

const (
	CmdBad = iota
	CmdNoOp
	CmdMove
)

type UserCmd struct {
	Msec         byte
	Buttons      byte
	Angles       client.Angles // viewangles
	MoveForward  int16
	MoveSideways int16
	MoveUp       int16
	Impulse      byte
	LightLevel   byte
}
