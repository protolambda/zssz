package dec

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type DecoderFn func(dr *DecodingReader, pointer unsafe.Pointer) error

type DecodingReader struct {
	input    io.Reader
	i        uint32
	max      uint32
	fuzzMode bool
}

func NewDecodingReader(input io.Reader) *DecodingReader {
	return &DecodingReader{input: input, i: 0, max: ^uint32(0)}
}

// returns a scope of the SSZ reader. Re-uses same scratchpad.
func (dr *DecodingReader) Scope(count uint32) (*DecodingReader, error) {
	if span := dr.GetBytesSpan(); span < count {
		return nil, fmt.Errorf("cannot create scoped decoding reader, scope of %d bytes is bigger than parent scope has available space %d", count, span)
	}
	return &DecodingReader{input: io.LimitReader(dr.input, int64(count)), i: 0, max: count}, nil
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

func (dr *DecodingReader) checkedIndexUpdate(x uint32) (n int, err error) {
	v := dr.i + x
	if v > dr.max {
		return int(dr.i), fmt.Errorf("cannot read %d bytes, %d beyond scope", x, v-dr.max)
	}
	dr.i = v
	return int(x), nil
}

func (dr *DecodingReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if n, err := dr.checkedIndexUpdate(uint32(len(p))); err != nil {
		return n, err
	}
	n := 0
	for n < len(p) {
		v, err := dr.input.Read(p[n:])
		n += v
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (dr *DecodingReader) ReadByte() (byte, error) {
	tmp := make([]byte, 1, 1)
	_, err := dr.Read(tmp)
	return tmp[0], err
}

func (dr *DecodingReader) ReadUint16() (uint16, error) {
	tmp := make([]byte, 2, 2)
	_, err := dr.Read(tmp)
	return binary.LittleEndian.Uint16(tmp), err
}

func (dr *DecodingReader) ReadUint32() (uint32, error) {
	tmp := make([]byte, 4, 4)
	_, err := dr.Read(tmp)
	return binary.LittleEndian.Uint32(tmp), err
}

func (dr *DecodingReader) ReadUint64() (uint64, error) {
	tmp := make([]byte, 8, 8)
	_, err := dr.Read(tmp)
	return binary.LittleEndian.Uint64(tmp), err
}

// returns the remaining span that can be read
func (dr *DecodingReader) GetBytesSpan() uint32 {
	return dr.Max() - dr.Index()
}

// if normal, offsets are used and enforced.
// if fuzzMode, no offsets are used, and lengths are read from the input, and adjusted to match remaining space.
func (dr *DecodingReader) IsFuzzMode() bool {
	return dr.fuzzMode
}
