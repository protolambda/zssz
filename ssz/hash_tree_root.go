package ssz

import (
	"unsafe"
)

type HashTreeRootFn func(HashFn, pointer unsafe.Pointer) []byte

func HashTreeRoot(hfn HashFn, val interface{}, sszTyp SSZ) []byte {
	p := unsafe.Pointer(&val)
	return sszTyp.HashTreeRoot(hfn, p)
}

func SigningRoot(hfn HashFn, val interface{}, sszTyp SignedSSZ) []byte {
	p := unsafe.Pointer(&val)
	return sszTyp.SigningRoot(hfn, p)
}

type HashFn func(input []byte) []byte
