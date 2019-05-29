package ssz

import "unsafe"

func EncodeSeries(elemSSZ SSZ, length uint32, fixedLen uint32,
	elemMemSize uintptr, eb *sszEncBuf, p unsafe.Pointer) {

	offset := uintptr(0)
	if elemSSZ.IsFixed() {
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + offset)
			offset += elemMemSize
			elemSSZ.Encode(eb, elemPtr)
		}
	} else {
		for i := uint32(0); i < length; i++ {
			elemPtr := unsafe.Pointer(uintptr(p) + offset)
			offset += elemMemSize
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
