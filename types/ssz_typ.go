package types

import (
	"unsafe"
	. "zssz/dec"
	. "zssz/enc"
	. "zssz/htr"
)

// Note: when this is changed,
//  don't forget to change the Read/PutUint32 calls that handle the length value in this allocated space.
const BYTES_PER_LENGTH_OFFSET = 4

type SSZ interface {
	// The length of the fixed-size part
	FixedLen() uint32
	// If the type is fixed-size
	IsFixed() bool
	// Reads object data from pointer, writes ssz-encoded data to EncodingBuffer
	Encode(eb *EncodingBuffer, p unsafe.Pointer)
	// Reads from input, populates object with read data
	Decode(dr *DecodingReader, p unsafe.Pointer) error
	HashTreeRoot(h *Hasher, pointer unsafe.Pointer) [32]byte
}

// SSZ definitions may also provide a way to compute a special hash-tree-root, for self-signed objects.
type SignedSSZ interface {
	SSZ
	SigningRoot(h *Hasher, p unsafe.Pointer) [32]byte
}
