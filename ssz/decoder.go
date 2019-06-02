package ssz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"runtime"
	"unsafe"
)

type DecoderFn func(dr *SSZDecReader, pointer unsafe.Pointer) error

func Decode(r io.Reader, val interface{}, sszTyp SSZ) error {
	dr := NewSSZDecReader(r)
	p := unsafe.Pointer(&val)
	err := sszTyp.Decode(dr, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return err
}

type SSZDecReader struct {
	input io.Reader
	scratch []byte
	i uint32
	max uint32
}

func NewSSZDecReader(input io.Reader) *SSZDecReader {
	return &SSZDecReader{input: input, scratch: make([]byte, 32, 32), i: 0, max: ^uint32(0)}
}

// returns a scope of the SSZ reader. Re-uses same scratchpad.
func (dr *SSZDecReader) Scope(count uint32) *SSZDecReader {
	return &SSZDecReader{input: io.LimitReader(dr.input, int64(count)), scratch: dr.scratch, i: 0, max: count}
}

// how far we have read so far (scoped per container)
func (dr *SSZDecReader) Index() uint32 {
	return dr.i
}

// How far we can read (max - i = remaining bytes to read without error).
// Note: when a child element is not fixed length,
// the parent should set the scope, so that the child can infer its size from it.
func (dr *SSZDecReader) Max() uint32 {
	return dr.max
}

func (dr *SSZDecReader) Read(p []byte) (n int, err error) {
	v := dr.i + uint32(len(p))
	if v > dr.max {
		return int(dr.i), fmt.Errorf("cannot read %d bytes, %d beyond scope", len(p), v - dr.max)
	}
	dr.i = v
	return dr.input.Read(p)
}

func (dr *SSZDecReader) ReadByte() (byte, error) {
	if dr.i + 1 > dr.max {
		return 0, errors.New("cannot read a single byte, it is beyond scope")
	}
	dr.i++
	_, err := dr.input.Read(dr.scratch[:1])
	return dr.scratch[0], err
}

func (dr *SSZDecReader) readBytes(count uint8) error {
	if count > 32 {
		return fmt.Errorf("cannot read more than 32 bytes into scratchpad")
	}
	_, err := dr.Read(dr.scratch[:count])
	return err
}

func (dr *SSZDecReader) readUint16() (uint16, error) {
	if err := dr.readBytes(2); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(dr.scratch[:2]), nil
}

func (dr *SSZDecReader) readUint32() (uint32, error) {
	if err := dr.readBytes(4); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(dr.scratch[:4]), nil
}

func (dr *SSZDecReader) readUint64() (uint64, error) {
	if err := dr.readBytes(8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(dr.scratch[:8]), nil
}
