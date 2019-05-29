package ssz

import "unsafe"

type DecoderFn func(pointer unsafe.Pointer)

func placeholderDecoder(p unsafe.Pointer) {
}
