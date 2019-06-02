package ssz

import (
	"encoding/binary"
	"runtime"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

type HashTreeRootFn func(hfn *Hasher, pointer unsafe.Pointer) [32]byte

func HashTreeRoot(h *Hasher, val interface{}, sszTyp SSZ) [32]byte {
	p := unsafe_util.IfacePtrToPtr(&val)
	out := sszTyp.HashTreeRoot(h, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return out
}

func SigningRoot(h *Hasher, val interface{}, sszTyp SignedSSZ) [32]byte {
	p := unsafe.Pointer(&val)
	return sszTyp.SigningRoot(h, p)
}

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
