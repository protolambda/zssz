package types

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"unsafe"
)

type SSZUint16 struct {}

func (t SSZUint16) FuzzMinLen() uint64 {
	return 2
}

func (t SSZUint16) FuzzMaxLen() uint64 {
	return 2
}

func (t SSZUint16) MinLen() uint64 {
	return 2
}

func (t SSZUint16) MaxLen() uint64 {
	return 2
}

func (t SSZUint16) FixedLen() uint64 {
	return 2
}

func (t SSZUint16) IsFixed() bool {
	return true
}

func (t SSZUint16) SizeOf(p unsafe.Pointer) uint64 {
	return 2
}

func (t SSZUint16) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	binary.LittleEndian.PutUint16(eb.Scratch[0:2], *(*uint16)(p))
	return eb.Write(eb.Scratch[0:2])
}

func (t SSZUint16) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	v, err := dr.ReadUint16()
	if err != nil {
		return err
	}
	*(*uint16)(p) = v
	return nil
}

func (t SSZUint16) DryCheck(dr *DecodingReader) error {
	_, err := dr.Skip(2)
	return err
}

func (t SSZUint16) HashTreeRoot(h MerkleFn, p unsafe.Pointer) (out [32]byte) {
	binary.LittleEndian.PutUint16(out[:], *(*uint16)(p))
	return
}

func (t SSZUint16) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	if *(*uint16)(p) == ^uint16(0) {
		w.Write("0xFFFF")
	} else {
		w.Write(fmt.Sprintf("%d", *(*uint16)(p)))
	}
}
