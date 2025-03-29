// These functions are wrappers for the regular Buffer read/write funcs for
// use with protobufs. Protos only support 32 and 64 bit signed/unsigned
// integers, but bytes, chars, shorts, words need to be read and written.
package message

func (b *Buffer) ReadCharP() int32 {
	return int32(b.ReadChar())
}

func (b *Buffer) ReadByteP() uint32 {
	return uint32(b.ReadByte())
}

func (b *Buffer) ReadShortP() int32 {
	return int32(b.ReadShort())
}

func (b *Buffer) ReadWordP() uint32 {
	return uint32(b.ReadWord())
}

func (b *Buffer) ReadLongP() int32 {
	return int32(b.ReadLong())
}

func (b *Buffer) WriteByteP(num uint32) {
	b.WriteByte(byte(num))
}

func (b *Buffer) WriteCharP(num int32) {
	b.WriteChar(uint8(num))
}

func (b *Buffer) WriteShortP(num uint32) {
	b.WriteShort(uint16(num))
}

func (b *Buffer) WriteWordP(num int32) {
	b.WriteWord(int16(num))
}

func (b *Buffer) WriteLongP(num int32) {
	b.WriteLong(int32(num))
}
