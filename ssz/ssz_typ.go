package ssz

import (
	"fmt"
	"reflect"
	"unsafe"
)

// Note: when this is changed,
//  don't forget to change the PutUint32 calls that make put the length in this allocated space.
const BYTES_PER_LENGTH_OFFSET = 4

type SSZ interface {
	// The length of the fixed-size part
	FixedLen() uint32
	// If the type is fixed-size
	IsFixed() bool
	// Reads object data from pointer, writes ssz-encoded data to sszEncBuf
	Encode(eb *sszEncBuf, p unsafe.Pointer)
	// Reads from input, populates object with read data
	Decode(p unsafe.Pointer)
	// Moves along input, ignores data, not populating any object
	Ignore()
}

func sszFactory(t reflect.Type) (SSZ, error) {
	switch t.Kind() {
	case reflect.Ptr:
		return sszFactory(t.Elem())
	case reflect.Bool:
		return sszBool, nil
	case reflect.Uint8:
		return sszUint8, nil
	case reflect.Uint16:
		return sszUint16, nil
	case reflect.Uint32:
		return sszUint32, nil
	case reflect.Uint64:
		return sszUint64, nil
	case reflect.Struct:
		return NewSSZContainer(t)
	case reflect.Array:
		elem_typ := t.Elem()
		switch elem_typ.Kind() {
		case reflect.Uint8:
			return NewSSZBytesN(elem_typ)
		default:
			// TODO: generic element type encoding
			return nil, fmt.Errorf("ssz: unrecognized array element type")
		}
	case reflect.Slice:
		elem_typ := t.Elem()
		switch elem_typ.Kind() {
		// TODO bytes (dynamic length) encoding
		default:
			// TODO: generic element type encoding
			return nil, fmt.Errorf("ssz: unrecognized array element type")
		}
	case reflect.String:
		// TODO string encoding
		return nil, nil
	default:
		return nil, fmt.Errorf("ssz: type %T cannot be serialized", t)
	}
}
