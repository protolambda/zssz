package ssz

import (
	"unsafe"
)

type HashTreeRootFn func(hfn *Hasher, pointer unsafe.Pointer) []byte

func HashTreeRoot(hfn *Hasher, val interface{}, sszTyp SSZ) []byte {
	p := unsafe.Pointer(&val)
	return sszTyp.HashTreeRoot(hfn, p)
}

func SigningRoot(hfn *Hasher, val interface{}, sszTyp SignedSSZ) []byte {
	p := unsafe.Pointer(&val)
	return sszTyp.SigningRoot(hfn, p)
}

type HashFn func(input []byte) []byte

type Hasher struct {
	Scratch []byte
	Hash HashFn
}
