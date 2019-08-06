package enc

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

type SizeFn func(pointer unsafe.Pointer) uint64

type EncoderFn func(eb *EncodingBuffer, pointer unsafe.Pointer)

type EncodingBuffer struct {
	buffer       bytes.Buffer
}

func (eb *EncodingBuffer) Bytes() []byte {
	return eb.buffer.Bytes()
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (eb *EncodingBuffer) Write(p []byte) {
	eb.buffer.Write(p)
}

// Write a single byte to the buffer.
func (eb *EncodingBuffer) WriteByte(v byte) {
	eb.buffer.WriteByte(v)
}

// Writes accumulated output in buffer to given writer.
func (eb *EncodingBuffer) WriteTo(w io.Writer) (n int64, err error) {
	return eb.buffer.WriteTo(w)
}

// Writes an offset for an element
func (eb *EncodingBuffer) WriteOffset(prevOffset uint64, elemLen uint64) uint64 {
	tmp := make([]byte, 4, 4)
	if prevOffset >= (uint64(1) << 32) {
		panic("cannot write offset with invalid previous offset")
	}
	if elemLen >= (uint64(1) << 32) {
		panic("cannot write offset with invalid element size")
	}
	offset := prevOffset + elemLen
	if offset >= (uint64(1) << 32) {
		panic("offset too large, not uint32")
	}
	binary.LittleEndian.PutUint32(tmp, uint32(offset))
	eb.buffer.Write(tmp)
	return offset
}

// Empties the buffers used
func (eb *EncodingBuffer) Reset() {
	eb.buffer.Reset()
}
