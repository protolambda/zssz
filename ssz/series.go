package ssz

import (
	"fmt"
	"unsafe"
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

func DecodeSeries(elemSSZ SSZ, length uint32, elemMemSize uintptr, dr *SSZDecReader, p unsafe.Pointer) error {
	memOffset := uintptr(0)
	if elemSSZ.IsFixed() {
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
			memOffset += elemMemSize
			if err := elemSSZ.Decode(dr, elemPtr); err != nil {
				return err
			}
		}
	} else {
		// technically we could also ignore offset correctness and skip ahead,
		//  but we may want to enforce proper offsets.
		offsets := make([]uint32, 0, length)
		startIndex := dr.Index()
		for i := uint32(0); i < length; i++ {
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
			elemPtr := unsafe.Pointer(uintptr(p) + memOffset)
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
