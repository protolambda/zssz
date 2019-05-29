package unsafe_util

import (
	"reflect"
	"unsafe"
)

// Allocates a new slice of the given length, with the given space for each element,
// and write a slice-header for it to the pointer.
func AllocateSliceSpaceAndBind(p unsafe.Pointer, length uint32, elemMemSize uintptr) unsafe.Pointer {
	dataLen := uint32(elemMemSize) * length
	data := make([]byte, 0, dataLen)
	contentsPtr := unsafe.Pointer(&data)
	sh := GetSliceHeader(contentsPtr, length)
	*(*reflect.SliceHeader)(p) = *sh
	return contentsPtr
}
