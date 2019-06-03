package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	"github.com/protolambda/zssz/util/ptrutil"
	"unsafe"
)

// pointer must point to start of the series contents
func EncodeVarSeries(encFn EncoderFn, length uint32, elemMemSize uintptr, eb *EncodingBuffer, p unsafe.Pointer) {
	memOffset := uintptr(0)
	fixedLen := BYTES_PER_LENGTH_OFFSET * length
	for i := uint32(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		// write an offset to the fixed data, to find the dynamic data with as a reader
		eb.WriteOffset(fixedLen)

		// encode the dynamic data to a temporary buffer
		temp := GetPooledBuffer()
		encFn(temp, elemPtr)
		// write it forward
		eb.WriteForward(temp)

		ReleasePooledBuffer(temp)
	}
	eb.FlushForward()
}

// pointer must point to start of the series contents
func decodeVarSeriesFromOffsets(decFn DecoderFn, offsets []uint32, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	length := uint32(len(offsets))
	var currentOffset uint32
	memOffset := uintptr(0)
	for i := uint32(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		// scope: until next offset, or end if this is the last item.
		currentOffset = dr.Index()
		if currentOffset != offsets[i] {
			return fmt.Errorf("expected to read to data %d bytes, got to %d", offsets[i], currentOffset)
		}
		var count uint32
		if i+1 < length {
			count = offsets[i+1] - currentOffset
		} else {
			count = dr.Max() - currentOffset
		}
		scoped := dr.Scope(count)
		if err := decFn(scoped, elemPtr); err != nil {
			return err
		}
		dr.UpdateIndexFromScoped(scoped)
	}
	if i, m := dr.Index(), dr.Max(); i != m {
		return fmt.Errorf("expected to finish reading the scope to max %d, got to %d", i, m)
	}
	return nil
}

// pointer must point to start of the series contents
func DecodeVarSeries(decFn DecoderFn, length uint32, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	// empty series are easy, always nothing to read.
	if length == 0 {
		return nil
	}

	// Read first offset, with this we can calculate the amount of expected offsets, i.e. the length of a slice.
	firstOffset, err := dr.ReadUint32()
	if err != nil {
		return err
	}

	if inducedLen := firstOffset / BYTES_PER_LENGTH_OFFSET; length != inducedLen {
		return fmt.Errorf("expected series of %d elements, got offset for %d elements", length, inducedLen)
	}

	// technically we could also ignore offset correctness and skip ahead,
	//  but we may want to enforce proper offsets.
	offsets := make([]uint32, 0, length)

	// add the first offset used in the length check
	offsets = append(offsets, firstOffset)

	// add the remaining offsets
	for i := uint32(1); i < length; i++ {
		offset, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		offsets = append(offsets, offset)
	}

	return decodeVarSeriesFromOffsets(decFn, offsets, elemMemSize, dr, p)
}

// pointer must point to the slice header to decode into
// (new space is allocated for contents and bound to the slice header when necessary)
func DecodeVarSlice(decFn DecoderFn, minElemLen uint32, bytesLen uint32, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	contentsPtr := p

	// empty series are easy, always nothing to read.
	if bytesLen == 0 {
		return nil
	}

	if startIndex := dr.Index(); startIndex != 0 {
		return fmt.Errorf("non-empty dynamic-length series has invalid starting index: %d", startIndex)
	}

	// Read first offset, with this we can calculate the amount of expected offsets, i.e. the length of a slice.
	firstOffset, err := dr.ReadUint32()
	if err != nil {
		return err
	}

	if firstOffset > bytesLen || (firstOffset%BYTES_PER_LENGTH_OFFSET) != 0 {
		return fmt.Errorf("non-empty dynamic-length series has invalid first offset: %d", firstOffset)
	}

	length := firstOffset / BYTES_PER_LENGTH_OFFSET

	if maxLen, minLen := uint64(dr.Max()), uint64(minElemLen)*uint64(length); minLen > maxLen {
		return fmt.Errorf("cannot fit %d elements of each a minimum size %d (%d total bytes) in %d bytes", length, minElemLen, minLen, maxLen)
	}

	// We don't want elements to be put in the slice header memory,
	// instead, we allocate the slice data, and change the contents-pointer in the header.
	contentsPtr = ptrutil.AllocateSliceSpaceAndBind(p, length, elemMemSize)

	offsets := make([]uint32, 0, length)

	// add the first offset used in the length check
	offsets = append(offsets, firstOffset)

	// add the remaining offsets
	for i := uint32(1); i < length; i++ {
		offset, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		offsets = append(offsets, offset)
	}

	if expectedIndex, currentIndex := BYTES_PER_LENGTH_OFFSET*length, dr.Index(); currentIndex != expectedIndex {
		return fmt.Errorf("expected to read to %d bytes, got to %d", expectedIndex, currentIndex)
	}

	return decodeVarSeriesFromOffsets(decFn, offsets, elemMemSize, dr, contentsPtr)
}
