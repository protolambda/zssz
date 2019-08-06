package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	"github.com/protolambda/zssz/util/ptrutil"
	"unsafe"
)

// pointer must point to start of the series contents
func EncodeVarSeries(encFn EncoderFn, sizeFn SizeFn, length uint64, elemMemSize uintptr, eb *EncodingBuffer, p unsafe.Pointer) {
	// the previous offset, to calculate a new offset from, starting after the fixed data.
	prevOffset := BYTES_PER_LENGTH_OFFSET * length
	// span of the previous var-size element
	prevSize := uint64(0)

	// first, write all the offsets
	memOffset := uintptr(0)
	for i := uint64(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize

		prevOffset = eb.WriteOffset(prevOffset, prevSize)
		prevSize = sizeFn(elemPtr)
	}

	// write all the data contents referenced by the offsets.
	memOffset = uintptr(0)
	for i := uint64(0); i < length; i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize

		encFn(eb, elemPtr)
	}
}

// pointer must point to start of the series contents
func decodeVarSeriesFromOffsets(decFn DecoderFn, offsets []uint64, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	for i := 0; i < len(offsets); i++ {
		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		currentOffset := dr.Index()
		if currentOffset != offsets[i] {
			return fmt.Errorf("expected to read to data %d bytes, got to %d", offsets[i], currentOffset)
		}
		// calculate the scope based on next offset, and max. value of this scope for the last value
		var scope uint64
		if next := i + 1; next < len(offsets) {
			if nextOffset := offsets[next]; nextOffset >= currentOffset {
				scope = nextOffset - currentOffset
			} else {
				return fmt.Errorf("offset %d is invalid", i)
			}
		} else {
			scope = dr.Max() - currentOffset
		}
		scoped, err := dr.Scope(scope)
		if err != nil {
			return err
		}
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
func DecodeVarSeries(decFn DecoderFn, length uint64, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	// empty series are easy, always nothing to read.
	if length == 0 {
		return nil
	}

	// Read first offset, with this we can calculate the amount of expected offsets, i.e. the length of a slice.
	firstOffset, err := dr.ReadOffset()
	if err != nil {
		return err
	}

	if derivedLen := firstOffset / BYTES_PER_LENGTH_OFFSET; length != derivedLen {
		return fmt.Errorf("expected series of %d elements, got offset for %d elements", length, derivedLen)
	}

	// technically we could also ignore offset correctness and skip ahead,
	//  but we may want to enforce proper offsets.
	offsets := make([]uint64, 0, length)

	// add the first offset used in the length check
	offsets = append(offsets, firstOffset)

	// add the remaining offsets
	for i := uint64(1); i < length; i++ {
		offset, err := dr.ReadOffset()
		if err != nil {
			return err
		}
		offsets = append(offsets, offset)
	}

	return decodeVarSeriesFromOffsets(decFn, offsets, elemMemSize, dr, p)
}

func DecodeVarSeriesFuzzMode(elem SSZ, length uint64, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	elemFuzzReqLen := elem.FuzzMinLen()
	lengthLeftOver := length * elemFuzzReqLen

	for i := uint64(0); i < length; i++ {
		lengthLeftOver -= elemFuzzReqLen
		span := dr.GetBytesSpan()
		if span < lengthLeftOver {
			return fmt.Errorf("under estimated length requirements for fuzzing input, not enough data available to fuzz")
		}
		available := span - lengthLeftOver

		scoped, err := dr.Scope(available)
		if err != nil {
			return err
		}
		scoped.EnableFuzzMode()

		elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
		memOffset += elemMemSize
		if err := elem.Decode(scoped, elemPtr); err != nil {
			return err
		}
		dr.UpdateIndexFromScoped(scoped)
	}
	return nil
}

// pointer must point to the slice header to decode into
// (new space is allocated for contents and bound to the slice header when necessary)
func DecodeVarSlice(decFn DecoderFn, minElemLen uint64, bytesLen uint64, limit uint64,
	alloc ptrutil.SliceAllocationFn, elemMemSize uintptr, dr *DecodingReader, p unsafe.Pointer) error {

	contentsPtr := p

	// empty series are easy, always nothing to read.
	if bytesLen == 0 {
		return nil
	}

	if startIndex := dr.Index(); startIndex != 0 {
		return fmt.Errorf("non-empty dynamic-length series has invalid starting index: %d", startIndex)
	}

	// Read first offset, with this we can calculate the amount of expected offsets, i.e. the length of a slice.
	firstOffset, err := dr.ReadOffset()
	if err != nil {
		return err
	}

	if firstOffset > bytesLen || (firstOffset%BYTES_PER_LENGTH_OFFSET) != 0 {
		return fmt.Errorf("non-empty dynamic-length series has invalid first offset: %d", firstOffset)
	}

	length := firstOffset / BYTES_PER_LENGTH_OFFSET

	if length > limit {
		return fmt.Errorf("got %d elements, expected no more than %d elements", length, limit)
	}

	if maxLen, minLen := uint64(dr.Max()), uint64(minElemLen)*uint64(length); minLen > maxLen {
		return fmt.Errorf("cannot fit %d elements of each a minimum size %d (%d total bytes) in %d bytes", length, minElemLen, minLen, maxLen)
	}

	// We don't want elements to be put in the slice header memory,
	// instead, we allocate the slice data, and change the contents-pointer in the header.
	contentsPtr = alloc(p, length)

	offsets := make([]uint64, 0, length)

	// add the first offset used in the length check
	offsets = append(offsets, firstOffset)

	// add the remaining offsets
	for i := uint64(1); i < length; i++ {
		offset, err := dr.ReadOffset()
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
