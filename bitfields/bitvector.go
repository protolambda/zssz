package bitfields

import "fmt"

// Bitvectors should have a pointer-receiver BitLen function to derive its fixed bit-length from.
type Bitvector interface {
	Bitfield
}

// Helper function to implement Bitvector with.
// Checks if b can have the given length n in bits.
// It checks if:
//  1. b has the same amount of bytes as necessary for n bits.
//  2. unused bits in b are 0
func BitvectorCheck(b []byte, n uint32) error {
	byteLen := uint32(len(b))
	if expected := (n + 7) >> 3; byteLen != expected {
		return fmt.Errorf("bitvector %b of %d bytes has not expected length in bytes %d", b, n, expected)
	}
	if n&7 == 0 {  // n is a multiple of 8, so last byte fits
		return nil
	}
	last := b[byteLen-1]
	// if is not a multiple of 8, check if it is big enough to hold any non-zero contents in the last byte.
	if (last >> (n&7)) != 0 {
		return fmt.Errorf("bitvector %b with last byte 0b%b has not expected %d bits in last byte", b, last, n)
	}
	return nil
}
