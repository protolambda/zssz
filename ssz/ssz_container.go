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

func NewSSZContainer(typ reflect.Type) (*SSZContainer, error) {
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
		fieldSSZ, err := sszFactory(field.Type)
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
			temp := getPooledBuffer()
			f.ssz.Encode(temp, unsafe.Pointer(uintptr(p)+f.memOffset))
			// write it forward
			eb.WriteForward(temp.Bytes())

			releasePooledBuffer(temp)
		}
	}
	// All the dynamic data is appended to the fixed data
	eb.FlushForward()
}

func (v *SSZContainer) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	if v.IsFixed() {
		for _, f := range v.Fields {
			// if the container is fixed length, all fields are
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
		i := 0
		for _, f := range v.Fields {
			if !f.ssz.IsFixed() {
				if fieldIndex := dr.Index(); pivotIndex + offsets[i] != fieldIndex {
					return fmt.Errorf("expected to read to %d bytes, got to %d", pivotIndex + offsets[i], fieldIndex)
				}
				i++
				if err := f.ssz.Decode(dr, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (v *SSZContainer) Ignore() {
	// TODO skip ahead Length bytes in input
}
