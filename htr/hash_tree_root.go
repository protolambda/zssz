package htr

import (
	"crypto/sha256"
	"encoding/binary"
	"unsafe"
)

type HashTreeRootFn func(hfn HashFn, pointer unsafe.Pointer) [32]byte

type HashFn func(input []byte) [32]byte

var ZeroHashes [][32]byte

// initialize the zero-hashes pre-computed data with the given hash-function.
func InitZeroHashes(hFn HashFn) {
	ZeroHashes = make([][32]byte, 32)
	v := [64]byte{}
	for i := 0; i < 31; i++ {
		copy(v[:32], ZeroHashes[i][:])
		copy(v[32:], ZeroHashes[i][:])
		ZeroHashes[i+1] = hFn(v[:])
	}
}

func init() {
	InitZeroHashes(sha256.Sum256)
}

func (h HashFn) Combi(a [32]byte, b [32]byte) [32]byte {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], b[:])
	return h(v[:])
}

func (h HashFn) MixIn(a [32]byte, i uint32) [32]byte {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], make([]byte, 32, 32))
	binary.LittleEndian.PutUint32(v[32:], i)
	return h(v[:])
}
