package ssz

import (
	"bytes"
	"encoding/binary"
	"io"
	"runtime"
	"sync"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

type EncoderFn func(eb *sszEncBuf, pointer unsafe.Pointer)

func Encode(w io.Writer, val interface{}, sszTyp SSZ) error {
	eb := bufferPool.Get().(*sszEncBuf)
	defer bufferPool.Put(eb)
	eb.reset()

	p := unsafe_util.IfacePtrToPtr(&val)
	sszTyp.Encode(eb, p)

	_, err := eb.ToWriter(w)

	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)

	return err
}

// TODO: maybe make buffer-pools ssz-type specific,
//  and initialize buffers with the known fixed-size of a type, and maybe plus some extra if not fixed length?

// get a cleaned buffer from the pool
func GetPooledBuffer() *sszEncBuf {
	eb := bufferPool.Get().(*sszEncBuf)
	eb.reset()
	return eb
}

func ReleasePooledBuffer(eb *sszEncBuf) {
	bufferPool.Put(eb)
}

// Encoding Buffers are pooled.
var bufferPool = sync.Pool{
	New: func() interface{} { return &sszEncBuf{} },
}

type sszEncBuf struct {
	buffer bytes.Buffer
	forward bytes.Buffer
	scratch [32]byte
}

func (eb *sszEncBuf) Bytes() []byte {
	return eb.Bytes()
}

// Writes to the forward buffer.
func (eb *sszEncBuf) WriteForward(p []byte) {
	eb.forward.Write(p)
}

// Writes the forward buffer to the main buffer, and resets the forward buffer.
func (eb *sszEncBuf) FlushForward() {
	eb.buffer.Write(eb.forward.Bytes())
	eb.forward.Reset()
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (eb *sszEncBuf) Write(p []byte) {
	eb.buffer.Write(p)
}

// Write a single byte to the buffer.
func (eb *sszEncBuf) WriteByte(v byte) {
	eb.buffer.WriteByte(v)
}

// Writes accumulated output in buffer to given writer.
func (eb *sszEncBuf) ToWriter(w io.Writer) (int64, error) {
	return eb.buffer.WriteTo(w)
}

// Writes an offset, calculated as len(forward) + fixedLen, to the end of the buffer
func (eb *sszEncBuf) WriteOffset(fixedLen uint32) {
	offset := uint32(eb.forward.Len())

	binary.LittleEndian.PutUint32(eb.scratch[:4], fixedLen+offset)
	eb.buffer.Write(eb.scratch[:4])
}

// Empties the buffers used
func (eb *sszEncBuf) reset() {
	eb.buffer.Reset()
	eb.forward.Reset()
}

