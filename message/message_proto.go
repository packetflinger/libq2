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
	b.WriteByte(int(num))
}

func (b *Buffer) WriteCharP(num int32) {
	b.WriteChar(int(num))
}

func (b *Buffer) WriteShortP(num int32) {
	b.WriteShort(int(num))
}

func (b *Buffer) WriteWordP(num uint32) {
	b.WriteWord(int(num))
}

func (b *Buffer) WriteLongP(num int32) {
	b.WriteLong(int(num))
}
