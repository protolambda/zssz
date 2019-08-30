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
	bitLimit  uint64
	byteLimit uint64 // exclusive delimiting bit
	leafLimit uint64 // exclusive delimiting bit
}

var bitlistMeta = reflect.TypeOf((*bitfields.BitlistMeta)(nil)).Elem()

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
	byteLimit := (bitLimit + 7) >> 3
	res := &SSZBitlist{
		bitLimit:  bitLimit,
		byteLimit: byteLimit,
		leafLimit: (byteLimit + 31) >> 5,
	}
	return res, nil
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) FuzzMinLen() uint64 {
	// 8 for a random byte count, 1 for a random leading byte
	return 8 + 1
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) FuzzMaxLen() uint64 {
	// 8 for a random byte count, limit for maximum fill
	return 8 + v.byteLimit
}

// in bytes (rounded up), not bits. Includes the delimiting 1 bit.
func (v *SSZBitlist) MinLen() uint64 {
	// leading bit to mark it the 0 length makes it 1 byte.
	return 1
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) MaxLen() uint64 {
	return (v.bitLimit >> 3) + 1
}

// in bytes (rounded up), not bits
func (v *SSZBitlist) FixedLen() uint64 {
	return 0
}

func (v *SSZBitlist) IsFixed() bool {
	return false
}

func (v *SSZBitlist) SizeOf(p unsafe.Pointer) uint64 {
	sh := ptrutil.ReadSliceHeader(p)
	return uint64(sh.Len)
}

func (v *SSZBitlist) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	return eb.Write(data)
}

func (v *SSZBitlist) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	var byteLen uint64
	if dr.IsFuzzMode() {
		x, err := dr.ReadUint64()
		if err != nil {
			return err
		}
		// get span to fill with available space
		span := dr.GetBytesSpan() - 1
		// respect type limit
		if span > v.byteLimit {
			span = v.byteLimit
		}
		if span != 0 {
			byteLen = x % span
		}
		// completely empty bitlists are invalid. Need a leading 1 bit.
		byteLen += 1
	} else {
		byteLen = dr.Max() - dr.Index()
	}
	// there may not be more bytes than necessary for the N bits, +1 for the delimiting bit.
	if byteLimitWithDelimiter := (v.bitLimit >> 3) + 1; byteLen > byteLimitWithDelimiter {
		return fmt.Errorf("got %d bytes, expected no more than %d bytes for bitlist",
			byteLen, byteLimitWithDelimiter)
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
	leaf := func(i uint64) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > byteLen {
			x := [32]byte{}
			copy(x[:], data[s:byteLen])
			if bitLen&7 != 0 && byteLen != 0 { // if we not already cut off the delimiting bit with a bytes boundary
				// find the index of the length-determining 1 bit (bitlist length == index of this bit)
				delimitByteIndex := (byteLen - 1) & 31
				mask := ^(byte(1) << bitfields.BitIndex(x[delimitByteIndex]))
				// zero out the length bit.
				x[delimitByteIndex] &= mask
			}
			return x[:]
		}
		// if the last leaf does not have to be padded,
		// then the length-determining bitlist bit is already cut off,
		// i.e. as sole bit in next (ignored) chunk of data.
		return data[s:e]
	}
	return h.MixIn(merkle.Merkleize(h, leafCount, v.leafLimit, leaf), bitLen)
}
