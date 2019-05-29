package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
	"zrnt-ssz/ssz/unsafe_util"
)

type SSZBytes struct {

}

func NewSSZBytes(typ reflect.Type) (*SSZBytes, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length bytes slice")
	}
	if typ.Elem().Kind() != reflect.Uint8 {
		return nil, fmt.Errorf("typ is not a bytes slice")
	}
	return &SSZBytes{}, nil
}

func (v *SSZBytes) FixedLen() uint32 {
	return 0
}

func (v *SSZBytes) IsFixed() bool {
	return true
}

func (v *SSZBytes) Encode(eb *sszEncBuf, p unsafe.Pointer) {
	sh := unsafe_util.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	eb.Write(data)
}

func (v *SSZBytes) Decode(dr *SSZDecReader, p unsafe.Pointer) error {
	sh := unsafe_util.ReadSliceHeader(p)
	data := *(*[]byte)(unsafe.Pointer(sh))
	_, err := dr.Read(data)
	return err
}

func (v *SSZBytes) HashTreeRoot(hFn HashFn, pointer unsafe.Pointer) []byte {

}
