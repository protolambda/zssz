package unsafe_util

import (
	"reflect"
	"unsafe"
)

func GetSliceHeader(p unsafe.Pointer, length uint32) *reflect.SliceHeader {
	return &reflect.SliceHeader{
		Data: uintptr(p),
		Len: int(length),
		Cap: int(length),
	}
}
