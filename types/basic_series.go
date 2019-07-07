package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/protolambda/zssz/util/ptrutil"
	"unsafe"
)

// WARNING: for little-endian architectures only, or the elem-length has to be 1 byte
func LittleEndianBasicSeriesEncode(eb *EncodingBuffer, p unsafe.Pointer, bytesLen uint32) {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))
	eb.Write(data)
}

// WARNING: for little-endian architectures only, or the elem-length has to be 1 byte
func LittleEndianBasicSeriesDecode(dr *DecodingReader, p unsafe.Pointer, bytesLen uint32, bytesLimit uint32, isBoolElem bool) error {
	if bytesLen > bytesLimit {
		return fmt.Errorf("got %d bytes, expected no more than %d bytes", bytesLen, bytesLimit)
	}
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))
	if _, err := dr.Read(data); err != nil {
		return err
	}
	if isBoolElem {
		if dr.IsFuzzMode() {
			// just make it correct where necessary
			for i := 0; i < len(data); i++ {
				if data[i] > 1 {
					data[i] = 1
				}
			}
		} else {
			for i := 0; i < len(data); i++ {
				if data[i] > 1 {
					return fmt.Errorf("byte %d in bool list is not a valid bool value: %d", i, data[i])
				}
			}
		}
	}
	return nil
}

// WARNING: for little-endian architectures only, or the elem-length has to be 1 byte
func LittleEndianBasicSeriesHTR(h HashFn, p unsafe.Pointer, bytesLen uint32, bytesLimit uint32) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))

	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		// pad the data
		if e > bytesLen {
			d := [32]byte{}
			copy(d[:], data[s:bytesLen])
			return d[:]
		}
		return data[s:e]
	}
	leafCount := (bytesLen + 31) >> 5
	leafLimit := (bytesLimit + 31) >> 5
	return merkle.Merkleize(h, leafCount, leafLimit, leaf)
}

func BigToLittleEndianChunk(data [32]byte, elemSize uint8) (out [32]byte) {
	// could be better with assembly or more bit-magic.
	// However, big-endian performance is not prioritized.
	x := 0
	for i := uint8(0); i < 32; i += elemSize {
		for j := elemSize - 1; j >= 1; j-- {
			out[x] = data[i|j]
			x++
		}
		out[x] = data[i]
		x++
	}
	return
}

// counter-part of LittleEndianBasicSeriesHTR
func BigEndianBasicSeriesHTR(h HashFn, p unsafe.Pointer, bytesLen uint32, bytesLimit uint32, elemSize uint8) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))

	leaf := func(i uint32) []byte {
		s := i << 5
		e := (i + 1) << 5
		d := [32]byte{}
		if e > bytesLen {
			copy(d[:], data[s:bytesLen])
		} else {
			copy(d[:], data[s:e])
		}
		d = BigToLittleEndianChunk(d, elemSize)
		return d[:]
	}
	leafCount := (bytesLen + 31) >> 5
	leafLimit := (bytesLimit + 31) >> 5
	return merkle.Merkleize(h, leafCount, leafLimit, leaf)
}
