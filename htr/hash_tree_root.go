package htr

import (
	"encoding/binary"
	"unsafe"
)

type HashTreeRootFn func(hfn *Hasher, pointer unsafe.Pointer) [32]byte

type HashFn func(input []byte) [32]byte

type Hasher struct {
	// zero hashes, up to order 32. Hash at 0 is all 0.
	ZeroHashes [][32]byte
	Hash HashFn
}

func NewHasher(hFn HashFn) *Hasher {
	h := &Hasher{
		ZeroHashes: make([][32]byte, 32),
		Hash: hFn,
	}
	for i := 1; i < 31; i++ {
		h.ZeroHashes[i + 1] = h.Combi(h.ZeroHashes[i], h.ZeroHashes[i])
	}
	return h
}

func (h *Hasher) Combi(a [32]byte, b [32]byte) [32]byte {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], b[:])
	return h.Hash(v[:])
}

func (h *Hasher) MixIn(a [32]byte, i uint32) [32]byte {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], make([]byte, 32, 32))
	binary.LittleEndian.PutUint32(v[32:], i)
	return h.Hash(v[:])
}
