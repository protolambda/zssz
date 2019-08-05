package types

import (
	"fmt"
	"github.com/protolambda/zssz/lists"
	"reflect"
)

var listType = reflect.TypeOf((*lists.List)(nil)).Elem()

func ReadListLimit(typ reflect.Type) (uint64, error) {
	ptrTyp := reflect.PtrTo(typ)
	if !ptrTyp.Implements(listType) {
		return 0, fmt.Errorf("*typ (pointer type) is not a ssz list")
	}
	typedNil := reflect.New(ptrTyp).Elem().Interface().(lists.List)
	return typedNil.Limit(), nil
}
