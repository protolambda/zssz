package bitfields

// Note: bitfield indices and lengths are generally all uint32, as this is used in SSZ for lengths too.

// General base interface for Bitlists and Bitvectors
// Note for Bitfields to work with the SSZ functionality:
//  - Bitlists need to be of kind []byte (packed bits, incl delimiter bit)
//  - Bitvectors need to be of kind [N]byte (packed bits)
type Bitfield interface {
	Get(i uint32) bool
	Set(i uint32, v bool)
}

// bitfields implementing this can be checked to be of a valid or not. Useful for untrusted bitfields.
// See BitlistCheck and BitvectorCheck to easily implement the validity checks.
type CheckedBitfield interface {
	Check() error
}

// the exact bitlength can be determined for bitfields implementing this method.
type SizedBits interface {
	BitLen() uint32
}

// Get index of left-most 1 bit.
// 0 (incl.) to 8 (excl.)
func BitIndex(v byte) (out uint32) {
	// going to be prettier with new Go 1.13 binary constant syntax
	if v&0xf0 != 0 { // 11110000
		out |= 4
		v >>= 4
	}
	if v&0x0c != 0 { // 00001100
		out |= 2
		v >>= 2
	}
	if v&0x02 != 0 { // 00000010
		out |= 1
		v >>= 1
	}
	return
}

// Helper function to implement Bitfields with.
// Assumes i is a valid bit-index to retrieve a bit from bytes b.
func GetBit(b []byte, i uint32) bool {
	return (b[i>>3]>>(i&7))&1 == 1
}

// Helper function to implement Bitfields with.
// Assumes i is a valid bit-index to set a bit within bytes b.
func SetBit(b []byte, i uint32, v bool) {
	if bit := byte(1) << (i & 7); v {
		b[i>>3] |= bit
	} else {
		b[i>>3] &^= bit
	}
}
