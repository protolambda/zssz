package ssz

import (
	"encoding/binary"
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

func (v *SSZBasic) Decode(p unsafe.Pointer) {
	v.Decoder(p)
}

func (v *SSZBasic) Ignore() {
	// TODO skip ahead Length bytes in input
}

var sszBool = &SSZBasic{1, func(eb *sszEncBuf, p unsafe.Pointer) {
	if *(*bool)(p) {
		eb.buffer.WriteByte(0x01)
	} else {
		eb.buffer.WriteByte(0x01)
	}
}, placeholderDecoder}

var sszUint8 = &SSZBasic{1, func(eb *sszEncBuf, p unsafe.Pointer) {
	eb.buffer.WriteByte(*(*byte)(p))
}, placeholderDecoder}

var sszUint16 = &SSZBasic{2, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint16(eb.NextBytes(2), *(*uint16)(p))
}, placeholderDecoder}

var sszUint32 = &SSZBasic{4, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint32(eb.NextBytes(4), *(*uint32)(p))
}, placeholderDecoder}

var sszUint64 = &SSZBasic{8, func(eb *sszEncBuf, p unsafe.Pointer) {
	binary.LittleEndian.PutUint64(eb.NextBytes(8), *(*uint64)(p))
}, placeholderDecoder}
