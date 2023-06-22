package client

const (
	ButtonAttack = 1
	ButtonUse    = 2
	ButtonAny    = 128
	ButtonMask   = ButtonAttack | ButtonUse | ButtonAny
	MoveAngle1   = 1 << 0
	MoveAngle2   = 1 << 1
	MoveAngle3   = 1 << 2
	MoveForward  = 1 << 3
	MoveSide     = 1 << 4
	MoveUp       = 1 << 5
	MoveButtons  = 1 << 6
	MoveImpulse  = 1 << 7
)

type ClientMove struct {
	Msec       byte
	Buttons    byte
	Angles     Angles
	Forward    int16
	Sideways   int16
	Up         int16
	Impulse    byte // needed?
	Lightlevel byte
}
