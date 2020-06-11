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

type SSZUint32 struct {}

func (t SSZUint32) FuzzMinLen() uint64 {
	return 4
}

func (t SSZUint32) FuzzMaxLen() uint64 {
	return 4
}

func (t SSZUint32) MinLen() uint64 {
	return 4
}

func (t SSZUint32) MaxLen() uint64 {
	return 4
}

func (t SSZUint32) FixedLen() uint64 {
	return 4
}

func (t SSZUint32) IsFixed() bool {
	return true
}

func (t SSZUint32) SizeOf(p unsafe.Pointer) uint64 {
	return 4
}

func (t SSZUint32) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	binary.LittleEndian.PutUint32(eb.Scratch[0:4], *(*uint32)(p))
	return eb.Write(eb.Scratch[0:4])
}

func (t SSZUint32) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	v, err := dr.ReadUint32()
	if err != nil {
		return err
	}
	*(*uint32)(p) = v
	return nil
}

func (t SSZUint32) DryCheck(dr *DecodingReader) error {
	_, err := dr.Skip(4)
	return err
}

func (t SSZUint32) HashTreeRoot(h MerkleFn, p unsafe.Pointer) (out [32]byte) {
	binary.LittleEndian.PutUint32(out[:], *(*uint32)(p))
	return
}

func (t SSZUint32) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	if *(*uint32)(p) == ^uint32(0) {
		w.Write("0xFFFFFFFF")
	} else {
		w.Write(fmt.Sprintf("%d", *(*uint32)(p)))
	}
}
