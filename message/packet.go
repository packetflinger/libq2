package message

import (
	"fmt"
	"net"
	"time"
)

type ClientPacket struct {
	Sequence1   int32
	Sequence2   int32
	QPort       uint16 //
	Reliable1   bool   // requires an ack
	Reliable2   bool
	MessageType byte   // what kind of msg?
	Data        []byte // the actual msg
}

// Out of band message
type ConnectionlessPacket struct {
	Sequence int32 // always -1
	Data     string
}

func NewConnectionlessPacket(str string) MessageBuffer {
	msg := MessageBuffer{}
	msg.WriteLong(-1)
	msg.WriteString(str)
	return msg
}

func NewClientCommand(str string) MessageBuffer {
	msg := MessageBuffer{}
	msg.WriteByte(CLCStringCommand)
	msg.WriteString(str)
	return msg
}

func (p ClientPacket) Marshal() []byte {
	msg := MessageBuffer{}
	msg.WriteLong(p.Sequence1)
	if p.Reliable1 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteLong(p.Sequence2)
	if p.Reliable2 {
		msg.Buffer[msg.Index-1] |= 0x80
	}
	msg.WriteShort(uint16(p.QPort))
	msg.WriteByte(p.MessageType)
	msg.WriteData(p.Data)
	return msg.Buffer
}

func (cp ConnectionlessPacket) Marshal() []byte {
	msg := MessageBuffer{}
	msg.WriteLong(-1)
	msg.WriteString(cp.Data)
	return msg.Buffer
}

func (cp ConnectionlessPacket) Send(srv string, port int) (MessageBuffer, error) {
	target := fmt.Sprintf("%s:%d", srv, port)

	// only use IPv4
	conn, err := net.Dial("udp4", target)
	if err != nil {
		return MessageBuffer{}, err
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Write(cp.Marshal())
	if err != nil {
		return MessageBuffer{}, err
	}

	d := make([]byte, 1500)
	read, err := conn.Read(d)
	if err != nil {
		// swallow read errors
		return MessageBuffer{}, nil
	}

	return MessageBuffer{Buffer: d[:read]}, nil
}
