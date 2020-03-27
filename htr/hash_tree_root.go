package htr

import (
	"crypto/sha256"
	"encoding/binary"
	"unsafe"
)

func init() {
	InitZeroHashes(NewHasherFunc(sha256.Sum256))
}

type HashTreeRootFn func(hfn Hasher, pointer unsafe.Pointer) [32]byte

type HashFn func(input []byte) [32]byte

type Hasher interface {
	Hash(a []byte) [32]byte
	Combi(a [32]byte, b [32]byte) [32]byte
	MixIn(a [32]byte, i uint64) [32]byte
}

type HasherFunc struct {
	b        [64]byte
	hashFunc HashFn
}

// Excluding the full zero bytes32 itself
const zeroHashesLevels = 64

var ZeroHashes [][32]byte

// initialize the zero-hashes pre-computed data with the given hash-function.
func InitZeroHashes(hFn Hasher) {
	ZeroHashes = make([][32]byte, zeroHashesLevels+1)
	v := [64]byte{}
	for i := 0; i < zeroHashesLevels; i++ {
		copy(v[:32], ZeroHashes[i][:])
		copy(v[32:], ZeroHashes[i][:])
		ZeroHashes[i+1] = hFn.Hash(v[:])
	}
}
func NewHasherFunc(h HashFn) *HasherFunc {
	return &HasherFunc{
		b:        [64]byte{},
		hashFunc: h,
	}
}
func (h *HasherFunc) Hash(a []byte) [32]byte {
	return h.hashFunc(a)
}

func (h *HasherFunc) Combi(a [32]byte, b [32]byte) [32]byte {
	copy(h.b[:32], a[:])
	copy(h.b[32:], b[:])
	return h.Hash(h.b[:])
}

func (h *HasherFunc) MixIn(a [32]byte, i uint64) [32]byte {
	copy(h.b[:32], a[:])
	copy(h.b[32:], make([]byte, 32, 32))
	binary.LittleEndian.PutUint64(h.b[32:], i)
	return h.Hash(h.b[:])
}
