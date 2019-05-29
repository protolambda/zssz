package unsafe_util

import (
	"unsafe"
)

// like reflect.SliceHeader. Unsafe.
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func GetSliceHeader(p unsafe.Pointer, length uint32) *SliceHeader {
	return &SliceHeader{
		Data: uintptr(p),
		Len: int(length),
		Cap: int(length),
	}
}
