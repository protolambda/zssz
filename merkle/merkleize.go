package merkle

import (
	. "github.com/protolambda/zssz/htr"
)

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
	// bitmagic: binary search through a uint32, offset down by 1 to not round powers of 2 up.
	// Then adding 1 to it to not get the index of the first bit, but the length of the bits (depth of tree)
	// Zero is a special case, it has a 0 depth.
	// Example:
	//  (in out): (0 0), (1 1), (2 1), (3 2), (4 2), (5 3), (6 3), (7 3), (8 3), (9 4)
	if v == 0 {
		return 0
	}
	v--
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
	out++
	return
}

// Merkleize with log(N) space allocation
func Merkleize(hasher HashFn, count uint32, limit uint32, leaf func(i uint32) []byte) (out [32]byte) {
	if count > limit {
		panic("merkleizing list that is too large, over limit")
	}
	if limit == 0 {
		return
	}
	if limit == 1 {
		copy(out[:], leaf(0))
		return
	}
	depth := GetDepth(count)
	limitDepth := GetDepth(limit)
	tmp := make([][32]byte, limitDepth+1, limitDepth+1)

	j := uint8(0)
	hArr := [32]byte{}
	h := hArr[:]

	merge := func(i uint32) {
		// merge back up from bottom to top, as far as we can
		for j = 0; ; j++ {
			// stop merging when we are in the left side of the next combi
			if i&(uint32(1)<<j) == 0 {
				// if we are at the count, we want to merge in zero-hashes for padding
				if i == count && j < depth {
					v := hasher.Combi(hArr, ZeroHashes[j])
					copy(h, v[:])
				} else {
					break
				}
			} else {
				// keep merging up if we are the right side
				v := hasher.Combi(tmp[j], hArr)
				copy(h, v[:])
			}
		}
		// store the merge result (may be no merge, i.e. bottom leaf node)
		copy(tmp[j][:], h)
	}

	// merge in leaf by leaf.
	for i := uint32(0); i < count; i++ {
		copy(h[:], leaf(i))
		merge(i)
	}

	// complement with 0 if empty, or if not the right power of 2
	if (uint32(1) << depth) != count {
		copy(h[:], ZeroHashes[0][:])
		merge(count)
	}

	// the next power of two may be smaller than the ultimate virtual size,
	// complement with zero-hashes at each depth.
	for j := depth; j < limitDepth; j++ {
		tmp[j+1] = hasher.Combi(tmp[j], ZeroHashes[j])
	}

	return tmp[limitDepth]
}
