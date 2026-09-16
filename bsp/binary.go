package bsp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// reader is a small sticky-error cursor used to decode the fixed-size
// binary records that make up a single .bsp lump. Once an error occurs,
// all further reads become no-ops so callers can defer error checking
// until after decoding a whole lump.
type reader struct {
	data []byte
	pos  int
	err  error
}

func newReader(data []byte) *reader {
	return &reader{data: data}
}

func (r *reader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if r.pos+n > len(r.data) {
		r.err = fmt.Errorf("unexpected end of lump data at offset %d, need %d more byte(s)", r.pos, n)
		return nil
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *reader) int16() int32 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return int32(int16(binary.LittleEndian.Uint16(b)))
}

func (r *reader) uint16() uint32 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return uint32(binary.LittleEndian.Uint16(b))
}

func (r *reader) int32Val() int32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(b))
}

func (r *reader) uint32Val() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func (r *reader) byteVal() uint32 {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return uint32(b[0])
}

func (r *reader) float32Val() float32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

// vector3 reads 3 consecutive floats.
func (r *reader) vector3() *bpb.Vector3 {
	x := r.float32Val()
	y := r.float32Val()
	z := r.float32Val()
	return &bpb.Vector3{X: x, Y: y, Z: z}
}

// shortBounds reads a mins/maxs pair stored as 6 signed 16-bit shorts, as
// used by dnode_t and dleaf_t. The values are widened into a float
// BoundingBox to fit the shared proto representation.
func (r *reader) shortBounds() *bpb.BoundingBox {
	mins := &bpb.Vector3{X: float32(r.int16()), Y: float32(r.int16()), Z: float32(r.int16())}
	maxs := &bpb.Vector3{X: float32(r.int16()), Y: float32(r.int16()), Z: float32(r.int16())}
	return &bpb.BoundingBox{Mins: mins, Maxs: maxs}
}

// floatBounds reads a mins/maxs pair stored as 6 floats, as used by dmodel_t.
func (r *reader) floatBounds() *bpb.BoundingBox {
	mins := r.vector3()
	maxs := r.vector3()
	return &bpb.BoundingBox{Mins: mins, Maxs: maxs}
}

// fixedString reads an n-byte, NUL-terminated/padded character array.
func (r *reader) fixedString(n int) string {
	b := r.take(n)
	if b == nil {
		return ""
	}
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}

// writer accumulates the fixed-size binary records for a single lump.
type writer struct {
	buf bytes.Buffer
}

func (w *writer) int16(v int32) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(int16(v)))
	w.buf.Write(b[:])
}

func (w *writer) uint16(v uint32) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	w.buf.Write(b[:])
}

func (w *writer) int32Val(v int32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(v))
	w.buf.Write(b[:])
}

func (w *writer) uint32Val(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf.Write(b[:])
}

func (w *writer) byteVal(v uint32) {
	w.buf.WriteByte(byte(v))
}

func (w *writer) float32Val(v float32) {
	w.uint32Val(math.Float32bits(v))
}

func (w *writer) vector3(v *bpb.Vector3) {
	w.float32Val(v.GetX())
	w.float32Val(v.GetY())
	w.float32Val(v.GetZ())
}

// shortBounds writes a mins/maxs pair as 6 signed 16-bit shorts, rounding
// each float to the nearest integer.
func (w *writer) shortBounds(b *bpb.BoundingBox) {
	mins, maxs := b.GetMins(), b.GetMaxs()
	w.int16(int32(math.Round(float64(mins.GetX()))))
	w.int16(int32(math.Round(float64(mins.GetY()))))
	w.int16(int32(math.Round(float64(mins.GetZ()))))
	w.int16(int32(math.Round(float64(maxs.GetX()))))
	w.int16(int32(math.Round(float64(maxs.GetY()))))
	w.int16(int32(math.Round(float64(maxs.GetZ()))))
}

// floatBounds writes a mins/maxs pair as 6 floats.
func (w *writer) floatBounds(b *bpb.BoundingBox) {
	w.vector3(b.GetMins())
	w.vector3(b.GetMaxs())
}

// fixedString writes s into an n-byte, zero-padded field. It errors if s
// (plus its implicit NUL terminator) doesn't fit.
func (w *writer) fixedString(s string, n int) error {
	if len(s) >= n {
		return fmt.Errorf("string %q is too long to fit in a %d-byte field", s, n)
	}
	b := make([]byte, n)
	copy(b, s)
	w.buf.Write(b)
	return nil
}

func (w *writer) bytes(b []byte) {
	w.buf.Write(b)
}

// pad appends zero bytes until the buffer length is a multiple of align.
func (w *writer) pad(align int) {
	if rem := w.buf.Len() % align; rem != 0 {
		w.buf.Write(make([]byte, align-rem))
	}
}
