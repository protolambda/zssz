package ssz

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
			return NewSSZBytesN(typ)
		// TODO: we could optimize by creating special basic-type vectors, like BytesN, for the other basic types
		default:
			return NewSSZVector(factory, typ)
		}
	case reflect.Slice:
		switch typ.Elem().Kind() {
		case reflect.Uint8:
			return NewSSZBytes(typ)
		default:
			return NewSSZList(factory, typ)
		}
	//case reflect.String:
	//	// TODO string encoding
	//	return nil, nil
	default:
		return nil, fmt.Errorf("ssz: type %T cannot be serialized", typ)
	}
}

