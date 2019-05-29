package ssz

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type SSZBasic struct {
	Length uint32
	Encoder EncoderFn
	Decoder DecoderFn
}

func (v *SSZBasic) FixedLen() uint32 {
	return v.Length
}

func (v *SSZBasic) IsFixed() bool {
	return true
}

func (v *SSZBasic) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	v.Encoder(eb, p)
}

func (v *SSZBasic) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	return v.Decoder(dr, p)
}

func (v *SSZBasic) HashTreeRoot(hFn HashFn, pointer unsafe.Pointer) []byte {

}

var sszBool = &SSZBasic{1, func(eb *sszEncBuf, p unsafe.Pointer) {
	if *(*bool)(p) {
		eb.buffer.WriteByte(0x01)
	} else {
		eb.buffer.WriteByte(0x01)
	}
}, func(dr *SSZDecReader, p unsafe.Pointer) error {
	b, err := dr.ReadByte()
	if err != nil {
		return err
	}
	if b == 0x00 {
		*(*bool)(p) = false
		return nil
	} else if b == 0x01 {
		*(*bool)(p) = true
		return nil
	} else {
		return fmt.Errorf("bool value is invalid")
	}
}}

var sszUint8 = &SSZBasic{1, func(eb *sszEncBuf, p unsafe.Pointer) {
	eb.buffer.WriteByte(*(*byte)(p))
}, func(dr *SSZDecReader, p unsafe.Pointer) error {
	b, err := dr.ReadByte()
	if err != nil {
		return err
	}
	*(*byte)(p) = b
	return nil
}}

var sszUint16 = &SSZBasic{2, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint16(eb.NextBytes(2), *(*uint16)(p))
}, func(dr *SSZDecReader, p unsafe.Pointer) error {
	v, err := dr.readUint16()
	if err != nil {
		return err
	}
	*(*uint16)(p) = v
	return nil
}}

var sszUint32 = &SSZBasic{4, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint32(eb.NextBytes(4), *(*uint32)(p))
}, func(dr *SSZDecReader, p unsafe.Pointer) error {
	v, err := dr.readUint32()
	if err != nil {
		return err
	}
	*(*uint32)(p) = binary.LittleEndian.Uint32(v)
	return nil
}}

var sszUint64 = &SSZBasic{8, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint64(eb.NextBytes(8), *(*uint64)(p))
}, func(dr *SSZDecReader, p unsafe.Pointer) error {
	v, err := dr.readUint64()
	if err != nil {
		return err
	}
	*(*uint64)(p) = binary.LittleEndian.Uint64(v)
	return nil
}}
