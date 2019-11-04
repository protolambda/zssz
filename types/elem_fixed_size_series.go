package types

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	"github.com/protolambda/zssz/util/ptrutil"
	"unsafe"
)

func EncodeFixedSeries(encFn EncoderFn, length uint64, elemMemSize uintptr, eb *EncodingWriter, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	for i := uint64(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		if err := encFn(eb, elemPtr); err != nil {
			return err
		}
	}
	return nil
}

func DecodeFixedSeries(decFn DecoderFn, length uint64, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	for i := uint64(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		if err := decFn(dr, elemPtr); err != nil {
			return err
		}
	}
	return nil
}

func VerifyFixedSeries(vFn VerifyFn, length uint64, dr *DecodingReader) error {
	for i := uint64(0); i < length; i++ {
		if err := vFn(dr); err != nil {
			return err
		}
	}
	return nil
}

func calcFixedSliceLength(elemLen uint64, bytesLen uint64, limit uint64) (uint64, error) {
	if elemLen == 0 {
		return 0, errors.New("cannot read a dynamic-length series of 0-length elements")
	}
	length := bytesLen / elemLen

	if length > limit {
		return 0, fmt.Errorf("got %d elements, expected no more than %d elements", length, limit)
	}
	return length, nil
}

func VerifyFixedSlice(vFn VerifyFn, elemLen uint64, bytesLen uint64, limit uint64, dr *DecodingReader) error {
	length, err := calcFixedSliceLength(elemLen, bytesLen, limit)
	if err != nil {
		return err
	}
	return VerifyFixedSeries(vFn, length, dr)
}

func DecodeFixedSlice(decFn DecoderFn, elemLen uint64, bytesLen uint64, limit uint64, alloc ptrutil.SliceAllocationFn, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	length, err := calcFixedSliceLength(elemLen, bytesLen, limit)
	if err != nil {
		return err
	}

	// it's a slice, we only have a header, we still need to allocate space for its data
	contentsPtr := alloc(p, length)

	return DecodeFixedSeries(decFn, length, elemMemSize, dr, contentsPtr)
}
