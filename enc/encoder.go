package enc

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

type EncoderFn func(eb *EncodingBuffer, pointer unsafe.Pointer)

type EncodingBuffer struct {
	buffer bytes.Buffer
	forward bytes.Buffer
}

func (eb *EncodingBuffer) Bytes() []byte {
	return eb.buffer.Bytes()
}

func (eb *EncodingBuffer) Read(p []byte) (n int, err error) {
	return eb.buffer.Read(p)
}

// Writes to the forward buffer.
func (eb *EncodingBuffer) WriteForward(data io.Reader) {
	_, err := eb.forward.ReadFrom(data)
	if err != nil {
		panic(err)
	}
}

// Writes the forward buffer to the main buffer, and resets the forward buffer.
func (eb *EncodingBuffer) FlushForward() {
	_, err := eb.buffer.ReadFrom(&eb.forward)
	if err != nil {
		panic(err)
	}
	eb.forward.Reset()
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

// Writes an offset, calculated as len(forward) + fixedLen, to the end of the buffer
func (eb *EncodingBuffer) WriteOffset(fixedLen uint32) {
	offset := uint32(eb.forward.Len())

	tmp := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(tmp, fixedLen+offset)
	eb.buffer.Write(tmp)
}

// Empties the buffers used
func (eb *EncodingBuffer) Reset() {
	eb.buffer.Reset()
	eb.forward.Reset()
}

