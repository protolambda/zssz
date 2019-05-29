package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

type SSZList struct {
	elemMemSize uintptr
	elemSSZ SSZ
}

func NewSSZList(factory SSZFactoryFn, typ reflect.Type) (*SSZList, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	elemTyp := typ.Elem()

	elemSSZ, err := factory(elemTyp)
	if err != nil {
		return nil, err
	}
	res := &SSZList{
		elemMemSize: elemTyp.Size(),
		elemSSZ: elemSSZ,
	}
	return res, nil
}

func (v *SSZList) FixedLen() uint32 {
	return 0
}

func (v *SSZList) IsFixed() bool {
	return false
}

func (v *SSZList) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	sh := unsafe_util.ReadSliceHeader(p)
	EncodeSeries(v.elemSSZ, uint32(sh.Len), v.elemMemSize, eb, p)
}

func (v *SSZList) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	sh := unsafe_util.ReadSliceHeader(p)
	return DecodeSeries(v.elemSSZ, uint32(sh.Len), v.elemMemSize, dr, p)
}
func (v *SSZList) HashTreeRoot(hFn HashFn, pointer unsafe.Pointer) []byte {

}
