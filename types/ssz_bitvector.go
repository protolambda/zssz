package types

import (
	"fmt"
	"github.com/protolambda/zssz/bitfields"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZBitvector struct {
	bitLen  uint32
	byteLen uint32
}

var bitvectorType = reflect.TypeOf((*bitfields.Bitvector)(nil)).Elem()

func NewSSZBitvector(typ reflect.Type) (*SSZBitvector, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length bytes array (bitvector requirement)")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes array (bitvector requirement)")
	}
	ptrTyp := reflect.PtrTo(typ)
	if !ptrTyp.Implements(bitvectorType) {
		return nil, fmt.Errorf("*typ (pointer type) is not a bitvector")
	}
	typedNil := reflect.New(ptrTyp).Elem().Interface().(bitfields.Bitvector)
	bitLen := typedNil.BitLen()
	byteLen := uint32(typ.Len())
	if (bitLen+7)>>3 != byteLen {
		return nil, fmt.Errorf("bitvector type has not the expected %d bytes to cover %d bits", byteLen, bitLen)
	}
	res := &SSZBitvector{bitLen: bitLen, byteLen: byteLen}
	return res, nil
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) FuzzReqLen() uint32 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) FixedLen() uint32 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) MinLen() uint32 {
	return v.byteLen
}

func (v *SSZBitvector) IsFixed() bool {
	return true
}

func (v *SSZBitvector) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBitvector) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	if _, err := dr.Read(data); err != nil {
		return err
	}
	// check if the data is a valid bitvector value (0 bits for unused bits)
	return bitfields.BitvectorCheck(data, v.bitLen)
}

func (v *SSZBitvector) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	leafCount := (v.byteLen + 31) >> 5
	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > v.byteLen {
			x := [32]byte{}
			copy(x[:], data[s:v.byteLen])
			return x[:]
		}
		return data[s:e]
	}
	return merkle.Merkleize(h, leafCount, leafCount, leaf)
}
