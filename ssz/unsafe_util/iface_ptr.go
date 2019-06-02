package unsafe_util

import "unsafe"

type iface struct {
	Type, Data unsafe.Pointer
}

func IfacePtrToPtr(val *interface{}) unsafe.Pointer {
	p := unsafe.Pointer(val)
	return (*iface)(p).Data
}
