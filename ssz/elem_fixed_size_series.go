package ssz

import (
	"errors"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

func EncodeFixedSeries(encFn EncoderFn, length uint32, elemMemSize uintptr, eb *sszEncBuf, p unsafe.Pointer) {
	memOffset := uintptr(0)
	for i := uint32(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		encFn(eb, elemPtr)
	}
}

// for dynamic-length series (Go slices), length is the amount of bytes available to read.
func DecodeFixedSeries(decFn DecoderFn, length uint32, elemMemSize uintptr, dr *SSZDecReader, p unsafe.Pointer) error {
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

func DecodeFixedSlice(decFn DecoderFn, elemLen uint32, bytesLen uint32, elemMemSize uintptr, dr *SSZDecReader, p unsafe.Pointer) error {
	if elemLen == 0 {
		return errors.New("cannot read a dynamic-length series of 0-length elements")
	}
	length := bytesLen / elemLen

	// it's a slice, we only have a header, we still need to allocate space for its data
	contentsPtr := unsafe_util.AllocateSliceSpaceAndBind(p, length, elemMemSize)

	return DecodeFixedSeries(decFn, length, elemMemSize, dr, contentsPtr)
}
