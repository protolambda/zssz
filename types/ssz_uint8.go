package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"unsafe"
)

type SSZUint8 struct{}

func (t SSZUint8) FuzzMinLen() uint64 {
	return 1
}

func (t SSZUint8) FuzzMaxLen() uint64 {
	return 1
}

func (t SSZUint8) MinLen() uint64 {
	return 1
}

func (t SSZUint8) MaxLen() uint64 {
	return 1
}

func (t SSZUint8) FixedLen() uint64 {
	return 1
}

func (t SSZUint8) IsFixed() bool {
	return true
}

func (t SSZUint8) SizeOf(p unsafe.Pointer) uint64 {
	return 1
}

func (t SSZUint8) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	return eb.WriteByte(*(*byte)(p))
}

func (t SSZUint8) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	b, err := dr.ReadByte()
	if err != nil {
		return err
	}
	*(*byte)(p) = b
	return nil
}

func (t SSZUint8) DryCheck(dr *DecodingReader) error {
	_, err := dr.Skip(1)
	return err
}

func (t SSZUint8) HashTreeRoot(h MerkleFn, p unsafe.Pointer) (out [32]byte) {
	out[0] = *(*byte)(p)
	return
}

func (t SSZUint8) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	w.Write(fmt.Sprintf("0x%02x", *(*byte)(p)))
}
