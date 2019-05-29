package ssz

import (
	"fmt"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

func EncodeSeries(elemSSZ SSZ, length uint32, elemMemSize uintptr, eb *sszEncBuf, p unsafe.Pointer) {
	memOffset := uintptr(0)
	if elemSSZ.IsFixed() {
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
			memOffset += elemMemSize
			elemSSZ.Encode(eb, elemPtr)
		}
	} else {
		fixedLen := BYTES_PER_LENGTH_OFFSET * length
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
			memOffset += elemMemSize
			// write an offset to the fixed data, to find the dynamic data with as a reader
			eb.WriteOffset(fixedLen)

			// encode the dynamic data to a temporary buffer
			temp := getPooledBuffer()
			elemSSZ.Encode(temp, elemPtr)
			// write it forward
			eb.WriteForward(temp.Bytes())

			releasePooledBuffer(temp)
		}
		eb.FlushForward()
	}
}

// for dynamic-length series (Go slices), length is the amount of bytes available to read.
func DecodeSeries(elemSSZ SSZ, length uint32, elemMemSize uintptr, dr *SSZDecReader, p unsafe.Pointer, isSlice bool) error {
	memOffset := uintptr(0)
	contentsPtr := p
	if elemSSZ.IsFixed() {
		elemLen := elemSSZ.FixedLen()
		if isSlice {
			if elemLen == 0 {
				return fmt.Errorf("cannot read a dynamic-length series of 0-length elements, of type %T", elemSSZ)
			} else {
				length = length / elemLen
			}
		}
		// if it's a slice, we still need to allocate space for it
		if isSlice {
			contentsPtr = unsafe_util.AllocateSliceSpaceAndBind(p, length, elemMemSize)
		}
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(contentsPtr) + memOffset)
			memOffset += elemMemSize
			if err := elemSSZ.Decode(dr, elemPtr); err != nil {
				return err
			}
		}
	} else {
		// empty series are easy, always nothing to read.
		if length == 0 {
			return nil
		}
		// technically we could also ignore offset correctness and skip ahead,
		//  but we may want to enforce proper offsets.
		var offsets []uint32
		if isSlice {
			// we don't know how many elements there are for now. Just grow the offsets as we read.
			offsets = make([]uint32, 0)
		} else {
			offsets = make([]uint32, 0, length)
		}
		startIndex := dr.Index()
		// Read first offset, with this we can calculate the amount of expected offsets, i.e. the length of a slice.
		firstOffset, err := dr.readUint32()
		if err != nil {
			return err
		}
		if isSlice {
			if startIndex != 0 {
				return fmt.Errorf("non-empty dynamic-length series has invalid starting index: %d", startIndex)
			}
			if firstOffset > length || ((firstOffset - startIndex) % BYTES_PER_LENGTH_OFFSET) != 0 {
				return fmt.Errorf("non-empty dynamic-length series has invalid first offset: %d", firstOffset)
			}
			length = (firstOffset - startIndex) / BYTES_PER_LENGTH_OFFSET
			// We don't want elements to be put in the slice header memory,
			// instead, we allocate the slice, and change the contents-pointer.
			contentsPtr = unsafe_util.AllocateSliceSpaceAndBind(p, length, elemMemSize)
		}
		// add the first offset, we need to check it later.
		offsets = append(offsets, firstOffset)
		// add the remaining offsets
		for i := uint32(1); i < length; i++ {
			offset, err := dr.readUint32()
			if err != nil {
				return err
			}
			offsets = append(offsets, offset)
		}
		pivotIndex := dr.Index()
		if expectedIndex := startIndex + (BYTES_PER_LENGTH_OFFSET * length); pivotIndex != expectedIndex {
			return fmt.Errorf("expected to read to %d bytes, got to %d", expectedIndex, pivotIndex)
		}
		var currentOffset uint32
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(contentsPtr) + memOffset)
			memOffset += elemMemSize
			// scope: until next offset, or end if this is the last item.
			currentOffset = dr.Index()
			if currentOffset != offsets[i] {
				return fmt.Errorf("expected to read to %d bytes, got to %d", offsets[i], currentOffset)
			}
			var count uint32
			if i + 1 < length {
				count = offsets[i + 1] - currentOffset
			} else {
				count = dr.Max() - currentOffset
			}
			scoped := dr.Scope(count)
			if err := elemSSZ.Decode(scoped, elemPtr); err != nil {
				return err
			}
		}
	}
	return nil
}
