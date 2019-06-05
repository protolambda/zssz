package htr

import (
	"crypto/sha256"
	"encoding/binary"
	"unsafe"
)

type HashTreeRootFn func(hfn HashFn, pointer unsafe.Pointer) [32]byte

type HashFn func(input []byte) [32]byte

var ZeroHashes [][32]byte

func init() {
	ZeroHashes = make([][32]byte, 32)
	hash := sha256.New()
	v := [64]byte{}
	for i := 1; i < 31; i++ {
		hash.Reset()
		copy(v[:32], ZeroHashes[i][:])
		copy(v[32:], ZeroHashes[i][:])
		hash.Write(v[:])
		copy(ZeroHashes[i + 1][:], hash.Sum(nil))
	}
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
