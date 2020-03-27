package types

import (
	"fmt"
	"github.com/protolambda/zssz/bitfields"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	. "github.com/protolambda/zssz/pretty"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZBitvector struct {
	bitLen  uint64
	byteLen uint64
}

var bitvectorMeta = reflect.TypeOf((*bitfields.BitvectorMeta)(nil)).Elem()

func NewSSZBitvector(typ reflect.Type) (*SSZBitvector, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a fixed-length bytes array (bitvector requirement)")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes array (bitvector requirement)")
	}
	ptrTyp := reflect.PtrTo(typ)
	if !ptrTyp.Implements(bitvectorMeta) {
		return nil, fmt.Errorf("*typ (pointer type) is not a bitvector")
	}
	typedNil := reflect.New(ptrTyp).Elem().Interface().(bitfields.BitvectorMeta)
	bitLen := typedNil.BitLen()
	byteLen := uint64(typ.Len())
	if (bitLen+7)>>3 != byteLen {
		return nil, fmt.Errorf("bitvector type has not the expected %d bytes to cover %d bits", byteLen, bitLen)
	}
	res := &SSZBitvector{bitLen: bitLen, byteLen: byteLen}
	return res, nil
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) FuzzMinLen() uint64 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) FuzzMaxLen() uint64 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) MinLen() uint64 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) MaxLen() uint64 {
	return v.byteLen
}

// in bytes (rounded up), not bits
func (v *SSZBitvector) FixedLen() uint64 {
	return v.byteLen
}

func (v *SSZBitvector) IsFixed() bool {
	return true
}

func (v *SSZBitvector) SizeOf(p unsafe.Pointer) uint64 {
	return v.byteLen
}

func (v *SSZBitvector) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	return eb.Write(data)
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

func (v *SSZBitvector) DryCheck(dr *DecodingReader) error {
	if v.bitLen == 0 {
		return nil
	}
	if v.byteLen > 1 {
		_, err := dr.Skip(v.byteLen - 1)
		if err != nil {
			return err
		}
	}
	last, err := dr.ReadByte()
	if err != nil {
		return err
	}
	return bitfields.BitvectorCheckLastByte(last, v.bitLen)
}

func (v *SSZBitvector) HashTreeRoot(h Hasher, p unsafe.Pointer) [32]byte {
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	leafCount := (v.byteLen + 31) >> 5
	leaf := func(i uint64) []byte {
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

func (v *SSZBitvector) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	sh := ptrutil.GetSliceHeader(p, v.byteLen)
	data := *(*[]byte)(unsafe.Pointer(sh))
	w.Write(fmt.Sprintf("%08b", data))
}
