package bitfields

import (
	"errors"
	"github.com/protolambda/zssz/lists"
)

type BitlistMeta interface {
	// Length (in bits) of the Bitlist.
	SizedBits
	// Limit (in bits) of the Bitlist.
	lists.List
}

type Bitlist interface {
	Bitfield
	BitlistMeta
}

// Returns the length of the bitlist.
// And although strictly speaking invalid. a sane default is returned for:
//  - an empty raw bitlist: a default 0 bitlist will be of length 0 too.
//  - a bitlist with a leading 0 byte: return the bitlist raw bit length,
//     excluding the last byte (As if it was full 0 padding).
func BitlistLen(b []byte) uint32 {
	byteLen := uint32(len(b))
	if byteLen == 0 {
		return 0
	}
	last := b[byteLen-1]
	return ((byteLen - 1) << 3) | BitIndex(last)
}

// Helper function to implement Bitlist with.
// Checks if b has the given length n in bits.
// It checks if:
//  0. the raw bitlist is not empty, there must be a 1 bit to determine the length.
//  1. the bitlist has a leading 1 bit in the last byte to determine the length with.
func BitlistCheck(b []byte) error {
	if len(b) == 0 {
		return errors.New("bitlist is missing length limit bit")
	}
	last := b[len(b)-1]
	if last == 0 {
		return errors.New("bitlist is invalid, trailing 0 byte")
	}
	return nil
}
