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

type SSZBitlist struct {
	bitLimit  uint32
	leafLimit uint32
}

var bitlistType = reflect.TypeOf((*bitfields.Bitlist)(nil)).Elem()

func NewSSZBitlist(typ reflect.Type) (*SSZBitlist, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length bytes slice (bitlist requirement)")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes slice (bitlist requirement)")
	}
	bitLimit, err := ReadListLimit(typ)
	if err != nil {
		return nil, err
	}
	res := &SSZBitlist{bitLimit: bitLimit, leafLimit: (((bitLimit + 7) >> 3) + 31) >> 5}
	return res, nil
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) FuzzReqLen() uint32 {
	// 4 for a random byte count, 1 for a random leading byte
	return 4 + 1
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) FixedLen() uint32 {
	return 0
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) MinLen() uint32 {
	// leading bit to mark it the 0 length makes it 1 byte.
	return 1
}

func (v *SSZBitlist) IsFixed() bool {
	return true
}

func (v *SSZBitlist) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBitlist) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	var byteLen uint32
	if dr.IsFuzzMode() {
		x, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		span := dr.GetBytesSpan() - 1
		if span != 0 {
			byteLen = x % span
		}
		// completely empty bitlists are invalid. Need a leading 1 bit.
		byteLen += 1
	} else {
		byteLen = dr.Max() - dr.Index()
	}
	// there may not be more bytes than necessary for the N bits, +1 for the delimiting bit.
	if byteLimit := ((v.bitLimit + 1) + 7) >> 3; byteLen > byteLimit {
		return fmt.Errorf("got %d bytes, expected no more than %d bytes to represent bitlist", byteLen, byteLimit)
	}
	ptrutil.BytesAllocFn(p, byteLen)
	data := *(*[]byte)(p)
	if _, err := dr.Read(data); err != nil {
		return err
	}
	if dr.IsFuzzMode() && len(data) > 1 && data[len(data)-1] == 0 {
		// last byte must not be 0 for bitlist to be valid
		data[len(data)-1] = 1
	}
	// check if the data is a valid bitvector value (0 bits for unused bits)
	return bitfields.BitlistCheck(data)
}

func (v *SSZBitlist) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	bitLen := bitfields.BitlistLen(data)
	byteLen := (bitLen + 7) >> 3
	leafCount := (byteLen + 31) >> 5
	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > byteLen {
			x := [32]byte{}
			copy(x[:], data[s:byteLen])
			// find the index of the length-determining 1 bit (bitlist length == index of this bit)
			chunkBitLen := bitfields.BitlistLen(x[:])
			bitfields.SetBit(x[:], chunkBitLen, false) // zero out the length bit.
			return x[:]
		}
		// if the last leaf does not have to be padded,
		// then the length-determining bitlist bit is already cut off,
		// i.e. as sole bit in next (ignored) chunk of data.
		return data[s:e]
	}
	return h.MixIn(merkle.Merkleize(h, leafCount, v.leafLimit, leaf), bitLen)
}
