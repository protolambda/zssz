package endianness

import (
	"unsafe"
)

var IsLittleEndian bool

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:
		IsLittleEndian = true
	case [2]byte{0xAB, 0xCD}:
		IsLittleEndian = false
	default:
		panic("Could not determine native endianness.")
	}
}
