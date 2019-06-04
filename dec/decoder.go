package dec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

type DecoderFn func(dr *DecodingReader, pointer unsafe.Pointer) error

type DecodingReader struct {
	input    io.Reader
	scratch  []byte
	i        uint32
	max      uint32
	fuzzMode bool
}

func NewDecodingReader(input io.Reader) *DecodingReader {
	return &DecodingReader{input: input, scratch: make([]byte, 32, 32), i: 0, max: ^uint32(0)}
}

// returns a scope of the SSZ reader. Re-uses same scratchpad.
func (dr *DecodingReader) Scope(count uint32) (*DecodingReader, error) {
	if span := dr.GetBytesSpan(); span < count {
		return nil, fmt.Errorf("cannot create scoped decoding reader, scope of %d bytes is bigger than parent scope has available space %d", count, span)
	}
	return &DecodingReader{input: io.LimitReader(dr.input, int64(count)), scratch: dr.scratch, i: 0, max: count}, nil
}

func (dr *DecodingReader) EnableFuzzMode() {
	dr.fuzzMode = true
}

func (dr *DecodingReader) UpdateIndexFromScoped(other *DecodingReader) {
	dr.i += other.i
}

// how far we have read so far (scoped per container)
func (dr *DecodingReader) Index() uint32 {
	return dr.i
}

// How far we can read (max - i = remaining bytes to read without error).
// Note: when a child element is not fixed length,
// the parent should set the scope, so that the child can infer its size from it.
func (dr *DecodingReader) Max() uint32 {
	return dr.max
}

func (dr *DecodingReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	v := dr.i + uint32(len(p))
	if v > dr.max {
		return int(dr.i), fmt.Errorf("cannot read %d bytes, %d beyond scope", len(p), v - dr.max)
	}
	dr.i = v
	return dr.input.Read(p)
}

func (dr *DecodingReader) ReadByte() (byte, error) {
	if dr.i + 1 > dr.max {
		return 0, errors.New("cannot read a single byte, it is beyond scope")
	}
	dr.i++
	_, err := dr.input.Read(dr.scratch[:1])
	return dr.scratch[0], err
}

func (dr *DecodingReader) ReadBytes(count uint8) error {
	if count > 32 {
		return fmt.Errorf("cannot read more than 32 bytes into scratchpad")
	}
	_, err := dr.Read(dr.scratch[:count])
	return err
}

func (dr *DecodingReader) ReadUint16() (uint16, error) {
	if err := dr.ReadBytes(2); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(dr.scratch[:2]), nil
}

func (dr *DecodingReader) ReadUint32() (uint32, error) {
	if err := dr.ReadBytes(4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dr.scratch[:4]), nil
}

func (dr *DecodingReader) ReadUint64() (uint64, error) {
	if err := dr.ReadBytes(8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(dr.scratch[:8]), nil
}

// returns the remaining span that can be read
func (dr *DecodingReader) GetBytesSpan() uint32 {
	return dr.Max() - dr.Index()
}

// if normal, offsets are used and enforced.
// if fuzzMode, no offsets are used, and lengths are read from the input, and adjusted to match remaining space.
func (dr *DecodingReader) IsRelaxed() bool {
	return dr.fuzzMode
}
