package ssz

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

type DecoderFn func(dr *SSZDecReader, pointer unsafe.Pointer) error

func placeholderDecoder(dr *SSZDecReader, p unsafe.Pointer) error {
	return nil
}

type ReadInput interface {
	io.Reader
	io.ByteReader
}

type SSZDecReader struct {
	input ReadInput
	scratch []byte
}

func NewSSZDecReader(input ReadInput) *SSZDecReader {
	return &SSZDecReader{input: input, scratch: make([]byte, 32, 32)}
}

func (dr *SSZDecReader) Read(p []byte) (n int, err error) {
	return dr.input.Read(p)
}

func (dr *SSZDecReader) ReadByte() (byte, error) {
	return dr.input.ReadByte()
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
