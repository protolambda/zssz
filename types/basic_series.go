package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
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
func LittleEndianBasicSeriesDecode(dr *DecodingReader, p unsafe.Pointer, bytesLen uint32, isBoolElem bool) error {
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
func LittleEndianBasicSeriesHTR(h *Hasher, p unsafe.Pointer, length uint32, bytesLen uint32, chunkPow uint8) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))
	dataLen := uint32(len(data))

	leaf := func(i uint32) []byte {
		s := i << chunkPow
		e := (i + 1) << chunkPow
		// pad the data
		if e > dataLen {
			d := [32]byte{}
			copy(d[:], data[s:dataLen])
			return d[:]
		}
		return data[s:e]
	}
	leafCount := (length + 31) >> 5
	return Merkleize(h, leafCount, leaf)
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
func BigEndianBasicSeriesHTR(h *Hasher, p unsafe.Pointer, length uint32, bytesLen uint32, chunkPow uint8) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))
	dataLen := uint32(len(data))

	elemSize := uint8(32) >> chunkPow

	leaf := func(i uint32) []byte {
		s := i << chunkPow
		e := (i + 1) << chunkPow
		d := [32]byte{}
		if e > dataLen {
			copy(d[:], data[s:dataLen])
		} else {
			copy(d[:], data[s:e])
		}
		d = BigToLittleEndianChunk(d, elemSize)
		return d[:]
	}
	leafCount := (length + 31) >> 5
	return Merkleize(h, leafCount, leaf)
}
