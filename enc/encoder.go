package enc

import (
	"encoding/binary"
	"io"
	"unsafe"
)

type SizeFn func(pointer unsafe.Pointer) uint64

type EncoderFn func(eb *EncodingWriter, pointer unsafe.Pointer) error

type EncodingWriter struct {
	w       io.Writer
	wfn     func(p []byte) (n int, err error)
	n       int
	Scratch [32]byte
}

func NewEncodingWriter(w io.Writer) *EncodingWriter {
	return &EncodingWriter{w: w, wfn: w.Write, n: 0}
}

// How many bytes were written to the underlying io.Writer before ending encoding (for handling errors)
func (ew *EncodingWriter) Written() int {
	return ew.n
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (ew *EncodingWriter) Write(p []byte) error {
	n, err := ew.wfn(p)
	ew.n += n
	return err
}

// Write a single byte to the buffer.
func (ew *EncodingWriter) WriteByte(v byte) error {
	ew.Scratch[0] = v
	return ew.Write(ew.Scratch[0:1])
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
	binary.LittleEndian.PutUint32(ew.Scratch[0:4], uint32(offset))
	err = ew.Write(ew.Scratch[0:4])
	return
}
