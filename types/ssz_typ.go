package types

import (
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"unsafe"
)

// Note: when this is changed,
//  don't forget to change the ReadOffset/WriteOffset calls that handle the length value in this allocated space.
const BYTES_PER_LENGTH_OFFSET = 4

type ChangeMode byte

const (
	Equal = iota
	Modified
	Added
	Deleted
)

type Change struct {
	Path        string
	Mode        ChangeMode
	Description string
}

type SSZFuzzInfo interface {
	// The minimum length to read the object from fuzzing mode
	FuzzMinLen() uint64
	// The maximum length to read the object from fuzzing mode
	FuzzMaxLen() uint64
}

type SSZInfo interface {
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
	// Verify the format of serialized data, without decoding the contents into memory
	Verify(dr *DecodingReader) error
}

type SSZMemory interface {
	// Gets the encoded size of the data under the given pointer.
	SizeOf(p unsafe.Pointer) uint64
	// Reads object data from pointer, writes ssz-encoded data to EncodingWriter
	// Returns (n, err), forwarded from the io.Writer being encoded into.
	Encode(eb *EncodingWriter, p unsafe.Pointer) error
	// Reads from input, populates object with read data
	Decode(dr *DecodingReader, p unsafe.Pointer) error
	// Hashes the object read at the given pointer
	HashTreeRoot(h HashFn, pointer unsafe.Pointer) [32]byte
	// Pretty print
	Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer)
	//// Diff two objects
	//Diff(a unsafe.Pointer, b unsafe.Pointer) []Change
}

type SSZ interface {
	SSZFuzzInfo
	SSZInfo
	SSZMemory
}

// SSZ definitions may also provide a way to compute a special hash-tree-root, for self-signed objects.
type SignedSSZ interface {
	SSZ
	SigningRoot(h HashFn, p unsafe.Pointer) [32]byte
}
