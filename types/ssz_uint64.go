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

type SSZUint64 struct {}

func (t SSZUint64) FuzzMinLen() uint64 {
	return 8
}

func (t SSZUint64) FuzzMaxLen() uint64 {
	return 8
}

func (t SSZUint64) MinLen() uint64 {
	return 8
}

func (t SSZUint64) MaxLen() uint64 {
	return 8
}

func (t SSZUint64) FixedLen() uint64 {
	return 8
}

func (t SSZUint64) IsFixed() bool {
	return true
}

func (t SSZUint64) SizeOf(p unsafe.Pointer) uint64 {
	return 8
}

func (t SSZUint64) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	v := [8]byte{}
	binary.LittleEndian.PutUint64(v[:], *(*uint64)(p))
	return eb.Write(v[:])
}

func (t SSZUint64) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	v, err := dr.ReadUint64()
	if err != nil {
		return err
	}
	*(*uint64)(p) = v
	return nil
}

func (t SSZUint64) DryCheck(dr *DecodingReader) error {
	_, err := dr.Skip(8)
	return err
}

func (t SSZUint64) HashTreeRoot(h HashFn, p unsafe.Pointer) (out [32]byte) {
	binary.LittleEndian.PutUint64(out[:], *(*uint64)(p))
	return
}

func (t SSZUint64) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	if *(*uint64)(p) == ^uint64(0) {
		w.Write("0xFFFFFFFFFFFFFFFF")
	} else {
		w.Write(fmt.Sprintf("%d", *(*uint64)(p)))
	}
}
