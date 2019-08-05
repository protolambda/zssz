package types

import (
	"fmt"
	"reflect"
)

type SSZFactoryFn func(typ reflect.Type) (SSZ, error)

// The default SSZ factory
func SSZFactory(typ reflect.Type) (SSZ, error) {
	return DefaultSSZFactory(SSZFactory, typ)
}

func DefaultSSZFactory(factory SSZFactoryFn, typ reflect.Type) (SSZ, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return NewSSZPtr(factory, typ)
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
		return NewSSZContainer(factory, typ)
	case reflect.Array:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			ptrTyp := reflect.PtrTo(typ)
			if ptrTyp.Implements(bitvectorMeta) {
				return NewSSZBitvector(typ)
			} else {
				return NewSSZBytesN(typ)
			}
		case reflect.Bool, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return NewSSZBasicVector(typ)
		default:
			return NewSSZVector(factory, typ)
		}
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			ptrTyp := reflect.PtrTo(typ)
			if ptrTyp.Implements(bitlistMeta) {
				return NewSSZBitlist(typ)
			} else {
				return NewSSZBytes(typ)
			}
		case reflect.Bool, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return NewSSZBasicList(typ)
		default:
			return NewSSZList(factory, typ)
		}
	// TODO: union, null, uint128, uint256, string
	default:
		return nil, fmt.Errorf("ssz: type %s cannot be recognized", typ.String())
	}
}
