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
	offset uintptr
	ssz    SSZ
}

type SSZContainer struct {
	Fields     []ContainerField
	isFixedLen bool
	fixedLen   uint32
}

func NewSSZContainer(typ reflect.Type) (*SSZContainer, error) {
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("typ is not a struct")
	}
	res := new(SSZContainer)
	res.isFixedLen = true
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
			res.isFixedLen = false
			res.fixedLen += BYTES_PER_LENGTH_OFFSET
		}
		res.Fields = append(res.Fields, ContainerField{offset: field.Offset, ssz: fieldSSZ})
	}
	return res, nil
}

func (v *SSZContainer) FixedLen() uint32 {
	return v.fixedLen
}

func (v *SSZContainer) IsFixed() bool {
	return v.isFixedLen
}

func (v *SSZContainer) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	u := uintptr(p)
	for _, f := range v.Fields {
		if f.ssz.IsFixed() {
			f.ssz.Encode(eb, unsafe.Pointer(u+f.offset))
		} else {
			// write an offset to the fixed data, to find the dynamic data with as a reader
			eb.WriteOffset(v.fixedLen)

			// encode the dynamic data to a temporary buffer
			temp := getPooledBuffer()
			f.ssz.Encode(temp, unsafe.Pointer(u+f.offset))
			// write it forward
			eb.WriteForward(temp.Bytes())

			releasePooledBuffer(temp)
		}
	}
	// All the dynamic data is appended to the fixed data
	eb.FlushForward()
}

func (v *SSZContainer) Decode(p unsafe.Pointer) {
	// TODO
}
func (v *SSZContainer) Ignore() {
	// TODO skip ahead Length bytes in input
}
