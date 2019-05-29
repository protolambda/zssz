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

func NewSSZList(typ reflect.Type) (*SSZList, error) {
	if typ.Kind() != reflect.Array {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	elemTyp := typ.Elem()

	elemSSZ, err := sszFactory(elemTyp)
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
	EncodeSeries(v.elemSSZ, uint32(sh.Len), 0, v.elemMemSize, eb, p)
}

func (v *SSZList) Decode(p unsafe.Pointer) {
	// TODO
}
func (v *SSZList) Ignore() {
	// TODO skip ahead Length bytes in input
}
