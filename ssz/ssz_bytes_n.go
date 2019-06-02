package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

type SSZBytesN struct {
	length uint32
}

func NewSSZBytesN(typ reflect.Type) (*SSZBytesN, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length bytes array")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes array")
	}
	length := typ.Len()
	res := &SSZBytesN{length: uint32(length)}
	return res, nil
}

func (v *SSZBytesN) VectorLength() uint32 {
	return v.length
}

func (v *SSZBytesN) FixedLen() uint32 {
	// 1 byte per element, just the same as the length
	return v.length
}

func (v *SSZBytesN) IsFixed() bool {
	return true
}

func (v *SSZBytesN) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	sh := unsafe_util.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBytesN) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	sh := unsafe_util.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytesN) HashTreeRoot(h *Hasher, p unsafe.Pointer) []byte {
	sh := unsafe_util.GetSliceHeader(p, v.length)
	data := *(*[]byte)(unsafe.Pointer(sh))
	dataLen := uint32(len(data))
	leafCount := (dataLen + 31) >> 5
	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the length
		if e > dataLen {
			h.ResetScratch32()
			copy(h.Scratch[:32], data[s:dataLen])
			return h.Scratch[:32]
		}
		return data[s:e]
	}
	return Merkleize(h, leafCount, leaf)
}
