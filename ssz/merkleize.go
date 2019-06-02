package ssz

const (
	mask0 = ^uint32((1 << (1 << iota)) - 1)
	mask1
	mask2
	mask3
	mask4
)

const (
	bit0 = uint8(1 << iota)
	bit1
	bit2
	bit3
	bit4
)

func GetDepth(v uint32) (out uint8) {
	// bitmagic: binary search through the uint32,
	// and set the corresponding index bit (5 bits -> 0x11111 = 31 is max.) in the output.
	if v&mask4 != 0 {
		v >>= bit4
		out |= bit4
	}
	if v&mask3 != 0 {
		v >>= bit3
		out |= bit3
	}
	if v&mask2 != 0 {
		v >>= bit2
		out |= bit2
	}
	if v&mask1 != 0 {
		v >>= bit1
		out |= bit1
	}
	if v&mask0 != 0 {
		out |= bit0
	}
	return
}

// Merkleize with log(N) space allocation
func Merkleize(hasher *Hasher, count uint32, leaf func(i uint32) []byte) []byte {
	if count == 0 {
		return make([]byte, 32)
	}
	if count == 1 {
		return leaf(0)
	}
	depth := GetDepth(count)
	tmp := make([][]byte, depth + 1, depth + 1)
	j := uint8(0)
	for i := uint32(0); i < count; i++ {
		h := leaf(i)
		// merge back up from bottom to top, as far as we can
		for j = 0; ; j++ {
			// stop merging when we are in the left side of the next combi
			if i & (uint32(1) << j) == 0 {
				break
			} else {
				// keep merging up if we are the right side
				h = hasher.Combi(tmp[j], h)
			}
		}
		// store the merge result (may be no merge, i.e. bottom leaf node)
		tmp[j] = h
	}
	// if count is not a power of 2
	if (count - 1) & count != 0 {
		i := count
		j = 0
		// walk up to the first right side
		for ; j < depth; j++ {
			if i & (uint32(1) << j) != 0 {
				break
			}
		}
		// the initial merge in is mixing in work from the right.
		// Initial work is the zero-hash at height j
		h := hasher.ZeroHashes[j]
		for ; j < depth; j++ {
			if i & (uint32(1) << j) == 0 {
				// left side: combine previous with zero-hash
				// i.e. venture out to merge back closer to the root
				h = hasher.Combi(h, hasher.ZeroHashes[j])
			} else {
				// right side: combine left side with zero hash
				// i.e. merge back with the work
				h = hasher.Combi(tmp[j], hasher.ZeroHashes[j])
			}
		}
		j--
		// store the merge result (may be no merge, i.e. bottom leaf node)
		tmp[j] = h
	}
	return tmp[j]
}
