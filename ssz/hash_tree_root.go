package ssz

import (
	"encoding/binary"
	"runtime"
	"unsafe"
)

type HashTreeRootFn func(hfn *Hasher, pointer unsafe.Pointer) []byte

func HashTreeRoot(h *Hasher, val interface{}, sszTyp SSZ) []byte {
	p := unsafe.Pointer(&val)
	out := sszTyp.HashTreeRoot(h, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return out
}

func SigningRoot(h *Hasher, val interface{}, sszTyp SignedSSZ) []byte {
	p := unsafe.Pointer(&val)
	return sszTyp.SigningRoot(h, p)
}

// Output slice must have a length of 32 bytes
type HashFn func(input []byte) []byte

type Hasher struct {
	// 64 bytes
	Scratch []byte
	// zero hashes, up to order 32. Hash at 0 is all 0.
	ZeroHashes [][]byte
	Hash HashFn
}

func NewHasher(hFn HashFn) *Hasher {
	h := &Hasher{
		Scratch: make([]byte, 32, 32),
		ZeroHashes: make([][]byte, 32),
		Hash: hFn,
	}
	h.ZeroHashes[0] = make([]byte, 32)
	for i := 1; i < 31; i++ {
		h.ZeroHashes[i + 1] = h.Combi(h.ZeroHashes[i], h.ZeroHashes[i])
	}
	return h
}

func (h *Hasher) ResetScratch64() {
	// 6 times faster than overwriting the slice with a new slice in a tight loop
	copy(h.Scratch, make([]byte, 64, 64))
}

func (h *Hasher) ResetScratch32() {
	copy(h.Scratch, make([]byte, 32, 32))
}

func (h *Hasher) Combi(a []byte, b []byte) []byte {
	copy(h.Scratch[:32], a)
	copy(h.Scratch[32:], b)
	return h.Hash(h.Scratch)
}

func (h *Hasher) MixIn(a []byte, v uint32) []byte {
	copy(h.Scratch[:32], a)
	copy(h.Scratch[32:], make([]byte, 32, 32))
	binary.LittleEndian.PutUint32(h.Scratch[32:], v)
	return h.Hash(h.Scratch)
}
