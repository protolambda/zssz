package enc

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type SizeFn func(pointer unsafe.Pointer) uint64

type EncoderFn func(eb *EncodingWriter, pointer unsafe.Pointer) error

type EncodingWriter struct {
	w io.Writer
	n int
}

func NewEncodingWriter(w io.Writer) *EncodingWriter {
	return &EncodingWriter{w: w, n: 0}
}

// How many bytes were written to the underlying io.Writer before ending encoding (for handling errors)
func (ew *EncodingWriter) Written() int {
	return ew.n
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (ew *EncodingWriter) Write(p []byte) error {
	n, err := ew.w.Write(p)
	ew.n += n
	return err
}

// Write a single byte to the buffer.
func (ew *EncodingWriter) WriteByte(v byte) error {
	return ew.Write([]byte{v})
}

// Writes an offset for an element
func (ew *EncodingWriter) WriteOffset(prevOffset uint64, elemLen uint64) (offset uint64, err error) {
	if prevOffset >= (uint64(1) << 32) {
		panic("cannot write offset with invalid previous offset")
	}
	if elemLen >= (uint64(1) << 32) {
		panic("cannot write offset with invalid element size")
	}
	offset = prevOffset + elemLen
	if offset >= (uint64(1) << 32) {
		panic("offset too large, not uint32")
	}
	tmp := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(offset))
	err = ew.Write(tmp)
	return
}
