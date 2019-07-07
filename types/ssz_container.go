package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/merkle"
	"github.com/protolambda/zssz/util/tags"
	"reflect"
	"unsafe"
)

const SSZ_TAG = "ssz"
const OMIT_FLAG = "omit"

type ContainerField struct {
	memOffset uintptr
	ssz       SSZ
	name      string
}

type SSZContainer struct {
	Fields      []ContainerField
	isFixedLen  bool
	fixedLen    uint32
	minLen      uint32
	offsetCount uint32
	fuzzReqLen  uint32
}

func NewSSZContainer(factory SSZFactoryFn, typ reflect.Type) (*SSZContainer, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("typ is not a struct")
	}
	res := new(SSZContainer)
	count := typ.NumField()
	for i := 0; i < count; i++ {
		field := typ.Field(i)
		if tags.HasFlag(&field, SSZ_TAG, OMIT_FLAG) {
			continue
		}
		fieldSSZ, err := factory(field.Type)
		if err != nil {
			return nil, err
		}
		if fieldSSZ.IsFixed() {
			res.fixedLen += fieldSSZ.FixedLen()
			res.minLen += fieldSSZ.MinLen()
		} else {
			res.fixedLen += BYTES_PER_LENGTH_OFFSET
			res.minLen += BYTES_PER_LENGTH_OFFSET + fieldSSZ.MinLen()
			res.offsetCount++
		}
		res.fuzzReqLen += fieldSSZ.FuzzReqLen()

		res.Fields = append(res.Fields, ContainerField{memOffset: field.Offset, ssz: fieldSSZ, name: field.Name})
	}
	res.isFixedLen = res.offsetCount == 0
	return res, nil
}

func (v *SSZContainer) FuzzReqLen() uint32 {
	return v.fuzzReqLen
}

func (v *SSZContainer) MinLen() uint32 {
	return v.minLen
}

func (v *SSZContainer) FixedLen() uint32 {
	return v.fixedLen
}

func (v *SSZContainer) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZContainer) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
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
			eb.WriteForward(temp)

			ReleasePooledBuffer(temp)
		}
	}
	// Only flush if we need to.
	// If not, forward can actually be filled with data from the parent container, and should not be flushed.
	if !v.IsFixed() {
		// All the dynamic data is appended to the fixed data
		eb.FlushForward()
	}
}

func (v *SSZContainer) decodeFixedSize(dr *DecodingReader, p unsafe.Pointer) error {
	for _, f := range v.Fields {
		// If the container is fixed length, all fields are.
		// No need to redefine the scope for fixed-length SSZ objects.
		if err := f.ssz.Decode(dr, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
			return err
		}
	}
	return nil
}

func (v *SSZContainer) decodeVarSizeFuzzmode(dr *DecodingReader, p unsafe.Pointer) error {
	lengthLeftOver := v.fuzzReqLen

	for _, f := range v.Fields {
		lengthLeftOver -= f.ssz.FuzzReqLen()
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
		if err := f.ssz.Decode(scoped, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
			return err
		}
		dr.UpdateIndexFromScoped(scoped)
	}
	return nil
}

func (v *SSZContainer) decodeVarSize(dr *DecodingReader, p unsafe.Pointer) error {
	// technically we could also ignore offset correctness and skip ahead,
	//  but we may want to enforce proper offsets.
	offsets := make([]uint32, 0, v.offsetCount)
	startIndex := dr.Index()
	fixedI := uint32(dr.Index())
	for _, f := range v.Fields {
		if f.ssz.IsFixed() {
			fixedI += f.ssz.FixedLen()
			// No need to redefine the scope for fixed-length SSZ objects.
			if err := f.ssz.Decode(dr, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
				return err
			}
		} else {
			fixedI += BYTES_PER_LENGTH_OFFSET
			// write an offset to the fixed data, to find the dynamic data with as a reader
			offset, err := dr.ReadUint32()
			if err != nil {
				return err
			}
			offsets = append(offsets, offset)
		}
		if i := dr.Index(); i != fixedI {
			return fmt.Errorf("fixed part had different size than expected, now at %d, expected to be at %d", i, fixedI)
		}
	}
	pivotIndex := dr.Index()
	if expectedIndex := v.fixedLen + startIndex; pivotIndex != expectedIndex {
		return fmt.Errorf("expected to read to %d bytes for fixed part of container, got to %d", expectedIndex, pivotIndex)
	}
	var currentOffset uint32
	i := 0
	for _, f := range v.Fields {
		if !f.ssz.IsFixed() {
			currentOffset = dr.Index()
			if offsets[i] != currentOffset {
				return fmt.Errorf("expected to read to %d bytes, got to %d", offsets[i], currentOffset)
			}
			// go to next offset
			i++
			// calculate the scope based on next offset, and max. value of this scope for the last value
			var count uint32
			if i < len(offsets) {
				if offset := offsets[i]; offset < currentOffset {
					return fmt.Errorf("offset %d is invalid", i)
				} else {
					count = offset - currentOffset
				}
			} else {
				count = dr.Max() - currentOffset
			}
			scoped, err := dr.Scope(count)
			if err != nil {
				return err
			}
			if err := f.ssz.Decode(scoped, unsafe.Pointer(uintptr(p)+f.memOffset)); err != nil {
				return err
			}
			dr.UpdateIndexFromScoped(scoped)
		}
	}
	return nil
}

func (v *SSZContainer) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if v.IsFixed() {
		return v.decodeFixedSize(dr, p)
	} else if dr.IsFuzzMode() {
		return v.decodeVarSizeFuzzmode(dr, p)
	} else {
		return v.decodeVarSize(dr, p)
	}
}

func (v *SSZContainer) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	leaf := func(i uint32) []byte {
		f := v.Fields[i]
		r := f.ssz.HashTreeRoot(h, unsafe.Pointer(uintptr(p)+f.memOffset))
		return r[:]
	}
	leafCount := uint32(len(v.Fields))
	return merkle.Merkleize(h, leafCount, leafCount, leaf)
}

func (v *SSZContainer) SigningRoot(h HashFn, p unsafe.Pointer) [32]byte {
	leaf := func(i uint32) []byte {
		f := v.Fields[i]
		r := f.ssz.HashTreeRoot(h, unsafe.Pointer(uintptr(p)+f.memOffset))
		return r[:]
	}
	// truncate last field
	leafCount := uint32(len(v.Fields))
	if leafCount != 0 {
		leafCount--
	}
	return merkle.Merkleize(h, leafCount, leafCount, leaf)
}
