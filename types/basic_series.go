package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

func GetBasicSSZElemType(kind reflect.Kind) (SSZ, error) {
	switch kind {
	case reflect.Bool:
		return SSZBool{}, nil
	case reflect.Uint8:
		return SSZUint8{}, nil
	case reflect.Uint16:
		return SSZUint16{}, nil
	case reflect.Uint32:
		return SSZUint32{}, nil
	case reflect.Uint64:
		return SSZUint64{}, nil
	default:
		return nil, fmt.Errorf("kind %d is not a basic type", kind)
	}
}

// WARNING: for little-endian architectures only, or the elem-length has to be 1 byte
func LittleEndianBasicSeriesEncode(eb *EncodingWriter, p unsafe.Pointer, bytesLen uint64) error {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))
	return eb.Write(data)
}

// WARNING: for little-endian architectures only, or the elem-length has to be 1 byte
func LittleEndianBasicSeriesDecode(dr *DecodingReader, p unsafe.Pointer, bytesLen uint64, bytesLimit uint64, isBoolElem bool) error {
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
func LittleEndianBasicSeriesHTR(h HashFn, p unsafe.Pointer, bytesLen uint64, bytesLimit uint64) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))

	leaf := func(i uint64) []byte {
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
func BigEndianBasicSeriesHTR(h HashFn, p unsafe.Pointer, bytesLen uint64, bytesLimit uint64, elemSize uint8) [32]byte {
	bytesSh := ptrutil.GetSliceHeader(p, bytesLen)
	data := *(*[]byte)(unsafe.Pointer(bytesSh))

	leaf := func(i uint64) []byte {
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

func CallSeries(fn func(i uint64, p unsafe.Pointer), length uint64, elemMemSize uintptr, p unsafe.Pointer) {
	memOffset := uintptr(0)
	for i := uint64(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		fn(i, elemPtr)
	}
}

func BasicSeriesDryCheck(dr *DecodingReader, bytesLen uint64, bytesLimit uint64, isBoolElem bool) error {
	if bytesLen > bytesLimit {
		return fmt.Errorf("got %d bytes, expected no more than %d bytes", bytesLen, bytesLimit)
	}
	if isBoolElem {
		for i := uint64(0); i < bytesLen; i++ {
			if v, err := dr.ReadByte(); err != nil {
				return err
			} else if v > 1 {
				return fmt.Errorf("byte %d in bool list is not a valid bool value: %d", i, v)
			}
		}
	} else {
		if _, err := dr.Skip(bytesLen); err != nil {
			return err
		}
	}
	return nil
}
