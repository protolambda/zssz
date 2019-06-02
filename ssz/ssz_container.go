package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
	"zrnt-ssz/ssz/tag_util"
)

const SSZ_TAG = "ssz"
const OMIT_FLAG = "omit"

type ContainerField struct {
	memOffset uintptr
	ssz       SSZ
}

type SSZContainer struct {
	Fields     []ContainerField
	isFixedLen bool
	fixedLen   uint32
	offsetCount uint32
}

func NewSSZContainer(factory SSZFactoryFn, typ reflect.Type) (*SSZContainer, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("typ is not a struct")
	}
	res := new(SSZContainer)
	count := typ.NumField()
	for i := 0; i < count; i++ {
		field := typ.Field(i)
		if tag_util.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
			continue
		}
		fieldSSZ, err := factory(field.Type)
		if err != nil {
			return nil, err
		}
		if fieldSSZ.IsFixed() {
			res.fixedLen += fieldSSZ.FixedLen()
		} else {
			res.fixedLen += BYTES_PER_LENGTH_OFFSET
			res.offsetCount++
		}
		res.Fields = append(res.Fields, ContainerField{memOffset: field.Offset, ssz: fieldSSZ})
	}
	res.isFixedLen = res.offsetCount == 0
	return res, nil
}

func (v *SSZContainer) FixedLen() uint32 {
	return v.fixedLen
}

func (v *SSZContainer) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZContainer) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	for _, f := range v.Fields {
		if f.ssz.IsFixed() {
			f.ssz.Encode(eb, unsafe.Pointer(uintptr(p)+f.memOffset))
		} else {
			// write an offset to the fixed data, to find the dynamic data with as a reader
			eb.WriteOffset(v.fixedLen)

			// encode the dynamic data to a temporary buffer
			temp := GetPooledBuffer()
			f.ssz.Encode(temp, unsafe.Pointer(uintptr(p)+f.memOffset))
			// write it forward
			eb.WriteForward(temp.Bytes())

			ReleasePooledBuffer(temp)
		}
	}
	// All the dynamic data is appended to the fixed data
	eb.FlushForward()
}

func (v *SSZContainer) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	if v.IsFixed() {
		for _, f := range v.Fields {
			// If the container is fixed length, all fields are.
			// No need to redefine the scope for fixed-length SSZ objects.
			if err := f.ssz.Decode(dr, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
				return err
			}
		}
	} else {
		// technically we could also ignore offset correctness and skip ahead,
		//  but we may want to enforce proper offsets.
		offsets := make([]uint32, 0, v.offsetCount)
		startIndex := dr.Index()
		for _, f := range v.Fields {
			if f.ssz.IsFixed() {
				// No need to redefine the scope for fixed-length SSZ objects.
				if err := f.ssz.Decode(dr, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
					return err
				}
			} else {
				// write an offset to the fixed data, to find the dynamic data with as a reader
				offset, err := dr.readUint32()
				if err != nil {
					return err
				}
				offsets = append(offsets, offset)
			}
		}
		pivotIndex := dr.Index()
		if expectedIndex := v.fixedLen + startIndex; pivotIndex != expectedIndex {
			return fmt.Errorf("expected to read to %d bytes, got to %d", expectedIndex, pivotIndex)
		}
		var currentOffset uint32
		for i, f := range v.Fields {
			if !f.ssz.IsFixed() {
				currentOffset = dr.Index()
				if offsets[i] != currentOffset {
					return fmt.Errorf("expected to read to %d bytes, got to %d", offsets[i], currentOffset)
				}
				// calculate the scope based on next offset, and max. value of this scope for the last value
				var count uint32
				if i + 1 < len(v.Fields) {
					count = offsets[i + 1] - currentOffset
				} else {
					count = dr.Max() - currentOffset
				}
				scoped := dr.Scope(count)
				if err := f.ssz.Decode(scoped, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *SSZContainer) HashTreeRoot(h *Hasher, p unsafe.Pointer) []byte {
	leaf := func(i uint32) []byte {
		f := v.Fields[i]
		return f.ssz.HashTreeRoot(h, unsafe.Pointer(uintptr(p)+f.memOffset))
	}
	return Merkleize(h, uint32(len(v.Fields)), leaf)
}

func (v *SSZContainer) SigningRoot(h *Hasher, p unsafe.Pointer) []byte {
	leaf := func(i uint32) []byte {
		f := v.Fields[i]
		return f.ssz.HashTreeRoot(h, unsafe.Pointer(uintptr(p)+f.memOffset))
	}
	// truncate last field
	leafCount := uint32(len(v.Fields))
	if leafCount != 0 {
		leafCount--
	}
	return Merkleize(h, leafCount, leaf)
}
