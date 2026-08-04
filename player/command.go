package player

import "github.com/packetflinger/libq2/message"

const (
	CM_ANGLE1  = 1 << 0
	CM_ANGLE2  = 1 << 1
	CM_ANGLE3  = 1 << 2
	CM_FORWARD = 1 << 3
	CM_SIDE    = 1 << 4
	CM_UP      = 1 << 5
	CM_BUTTONS = 1 << 6
	CM_IMPULSE = 1 << 7

	ButtonMask    = (1 << 0) | (1 << 1) | (1 << 7)
	ButtonForward = 1 << 2
	ButtonSide    = 1 << 3
	ButtonUp      = 1 << 4
)

// UserCommand is our usercmd_t implementation
type UserCommand struct {
	Msec        byte
	Buttons     byte
	Angles      [3]int16
	ForwardMove int16
	SideMove    int16
	UpMove      int16
	Impulse     byte
	LightLevel  byte
}

func (u UserCommand) WriteDeltaUsercmd(from UserCommand) message.Buffer {
	msg := message.Buffer{}
	bits := 0
	if u.Angles[0] != from.Angles[0] {
		bits |= CM_ANGLE1
	}
	if u.Angles[1] != from.Angles[1] {
		bits |= CM_ANGLE2
	}
	if u.Angles[2] != from.Angles[2] {
		bits |= CM_ANGLE3
	}
	if u.ForwardMove != from.ForwardMove {
		bits |= CM_FORWARD
	}
	if u.SideMove != from.SideMove {
		bits |= CM_SIDE
	}
	if u.UpMove != from.UpMove {
		bits |= CM_UP
	}
	if u.Buttons != from.Buttons {
		bits |= CM_BUTTONS
	}
	if u.Impulse != from.Impulse {
		bits |= CM_IMPULSE
	}
	msg.WriteByte(bits)
	buttons := u.Buttons & ButtonMask
	if (bits & CM_ANGLE1) > 0 {
		msg.WriteShort(int(u.Angles[0]))
	}
	if (bits & CM_ANGLE2) > 0 {
		msg.WriteShort(int(u.Angles[1]))
	}
	if (bits & CM_ANGLE3) > 0 {
		msg.WriteShort(int(u.Angles[2]))
	}
	if (bits & CM_FORWARD) > 0 {
		if (buttons & ButtonForward) > 0 {
			msg.WriteChar(int(u.ForwardMove) / 5)
		} else {
			msg.WriteShort(int(u.ForwardMove))
		}
	}
	if (bits & CM_SIDE) > 0 {
		if (buttons & ButtonSide) > 0 {
			msg.WriteChar(int(u.SideMove) / 5)
		} else {
			msg.WriteShort(int(u.SideMove))
		}
	}
	if (bits & CM_UP) > 0 {
		if (buttons & ButtonUp) > 0 {
			msg.WriteChar(int(u.UpMove) / 5)
		} else {
			msg.WriteShort(int(u.UpMove))
		}
	}
	if (bits & CM_BUTTONS) > 0 {
		msg.WriteByte(int(u.Buttons))
	}
	if (bits & CM_IMPULSE) > 0 {
		msg.WriteByte(int(u.Impulse))
	}
	msg.WriteByte(int(u.LightLevel))
	msg.WriteByte(int(u.Msec))
	return msg
}
