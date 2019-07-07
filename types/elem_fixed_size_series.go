package types

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	"github.com/protolambda/zssz/util/ptrutil"
	"unsafe"
)

func EncodeFixedSeries(encFn EncoderFn, length uint32, elemMemSize uintptr, eb *EncodingBuffer, p unsafe.Pointer) {
	memOffset := uintptr(0)
	for i := uint32(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		encFn(eb, elemPtr)
	}
}

func DecodeFixedSeries(decFn DecoderFn, length uint32, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	for i := uint32(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		if err := decFn(dr, elemPtr); err != nil {
			return err
		}
	}
	return nil
}

func DecodeFixedSlice(decFn DecoderFn, elemLen uint32, bytesLen uint32, limit uint32, alloc ptrutil.SliceAllocationFn, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	if elemLen == 0 {
		return errors.New("cannot read a dynamic-length series of 0-length elements")
	}
	length := bytesLen / elemLen

	if length > limit {
		return fmt.Errorf("got %d elements, expected no more than %d elements", length, limit)
	}

	// it's a slice, we only have a header, we still need to allocate space for its data
	contentsPtr := alloc(p, length)

	return DecodeFixedSeries(decFn, length, elemMemSize, dr, contentsPtr)
}
