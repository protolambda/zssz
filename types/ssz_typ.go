package types

import (
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"unsafe"
)

// Note: when this is changed,
//  don't forget to change the ReadOffset/WriteOffset calls that handle the length value in this allocated space.
const BYTES_PER_LENGTH_OFFSET = 4

type SSZ interface {
	// The minimum length to read the object from fuzzing mode
	FuzzMinLen() uint64
	// The maximum length to read the object from fuzzing mode
	FuzzMaxLen() uint64
	// The minimum length of the object.
	// If the object is fixed-len, this should equal FixedLen()
	MinLen() uint64
	// The maximum length of the object.
	// If the object is fixed-len, this should equal FixedLen()
	MaxLen() uint64
	// The length of the fixed-size part
	FixedLen() uint64
	// If the type is fixed-size
	IsFixed() bool
	// Reads object data from pointer, writes ssz-encoded data to EncodingBuffer
	Encode(eb *EncodingBuffer, p unsafe.Pointer)
	// Reads from input, populates object with read data
	Decode(dr *DecodingReader, p unsafe.Pointer) error
	// Hashes the object read at the given pointer
	HashTreeRoot(h HashFn, pointer unsafe.Pointer) [32]byte
}

// SSZ definitions may also provide a way to compute a special hash-tree-root, for self-signed objects.
type SignedSSZ interface {
	SSZ
	SigningRoot(h HashFn, p unsafe.Pointer) [32]byte
}
