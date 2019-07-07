package ptrutil

import (
	"unsafe"
)

type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func GetSliceHeader(p unsafe.Pointer, length uint32) *SliceHeader {
	return &SliceHeader{
		Data: p,
		Len:  int(length),
		Cap:  int(length),
	}
}

func ReadSliceHeader(p unsafe.Pointer) *SliceHeader {
	return (*SliceHeader)(p)
}
