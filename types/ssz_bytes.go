package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZBytes struct {
}

func NewSSZBytes(typ reflect.Type) (*SSZBytes, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length bytes slice")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes slice")
	}
	return &SSZBytes{}, nil
}

func (v *SSZBytes) FixedLen() uint32 {
	return 0
}

func (v *SSZBytes) MinLen() uint32 {
	return 0
}

func (v *SSZBytes) IsFixed() bool {
	return false
}

func (v *SSZBytes) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBytes) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	var length uint32
	if dr.IsFuzzMode() {
		x, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		length = x % dr.GetBytesSpan()
	} else {
		length = dr.Max() - dr.Index()
	}
	ptrutil.AllocateSliceSpaceAndBind(p, length, 1)
	data := *(*[]byte)(p)
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytes) HashTreeRoot(h *Hasher, p unsafe.Pointer) [32]byte {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	dataLen := uint32(len(data))
	leafCount := (dataLen + 31) >> 5
	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > dataLen {
			v := [32]byte{}
			copy(v[:], data[s:dataLen])
			return v[:]
		}
		return data[s:e]
	}
	return Merkleize(h, leafCount, leaf)
}
