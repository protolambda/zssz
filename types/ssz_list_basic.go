package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"github.com/protolambda/zssz/util/endianness"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZBasicList struct {
	alloc     ptrutil.SliceAllocationFn
	elemKind  reflect.Kind
	elemSSZ   SSZ
	limit     uint64
	byteLimit uint64
}

func NewSSZBasicList(typ reflect.Type) (*SSZBasicList, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	limit, err := ReadListLimit(typ)
	if err != nil {
		return nil, err
	}

	elemTyp := typ.Elem()
	elemKind := elemTyp.Kind()
	elemSSZ, err := GetBasicSSZElemType(elemKind)
	if err != nil {
		return nil, err
	}
	if elemSSZ.FixedLen() != uint64(elemTyp.Size()) {
		return nil, fmt.Errorf("basic element type has different size than SSZ type unexpectedly, ssz: %d, go: %d", elemSSZ.FixedLen(), elemTyp.Size())
	}

	res := &SSZBasicList{
		alloc:     ptrutil.MakeSliceAllocFn(typ),
		elemKind:  elemKind,
		elemSSZ:   elemSSZ,
		limit:     limit,
		byteLimit: limit * elemSSZ.FixedLen(),
	}
	return res, nil
}

func (v *SSZBasicList) FuzzMinLen() uint64 {
	return 8
}

func (v *SSZBasicList) FuzzMaxLen() uint64 {
	return 8 + v.byteLimit
}

func (v *SSZBasicList) MinLen() uint64 {
	return 0
}

func (v *SSZBasicList) MaxLen() uint64 {
	return v.byteLimit
}

func (v *SSZBasicList) FixedLen() uint64 {
	return 0
}

func (v *SSZBasicList) IsFixed() bool {
	return false
}

func (v *SSZBasicList) SizeOf(p unsafe.Pointer) uint64 {
	sh := ptrutil.ReadSliceHeader(p)
	return uint64(sh.Len) * v.elemSSZ.FixedLen()
}

func (v *SSZBasicList) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.ReadSliceHeader(p)

	// we can just write the data as-is in a few contexts:
	// - if we're in a little endian architecture
	// - if there is no endianness to deal with
	if endianness.IsLittleEndian || v.elemSSZ.FixedLen() == 1 {
		return LittleEndianBasicSeriesEncode(eb, sh.Data, uint64(sh.Len)*v.elemSSZ.FixedLen())
	} else {
		return EncodeFixedSeries(v.elemSSZ.Encode, uint64(sh.Len), uintptr(v.elemSSZ.FixedLen()), eb, sh.Data)
	}
}

func (v *SSZBasicList) decodeFuzzmode(dr *DecodingReader, p unsafe.Pointer) error {
	x, err := dr.ReadUint64()
	if err != nil {
		return err
	}
	span := dr.GetBytesSpan()
	if v.byteLimit > span {
		span = v.byteLimit
	}
	if span == 0 {
		return nil
	}
	bytesLen := x % span
	bytesLen -= bytesLen % v.elemSSZ.FixedLen()

	if endianness.IsLittleEndian || v.elemSSZ.FixedLen() == 1 {
		contentsPtr := v.alloc.MutateLenOrAllocNew(p, bytesLen/v.elemSSZ.FixedLen())
		bytesLimit := v.limit * v.elemSSZ.FixedLen()
		return LittleEndianBasicSeriesDecode(dr, contentsPtr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.limit, v.alloc, uintptr(v.elemSSZ.FixedLen()), dr, p)
	}
}

func (v *SSZBasicList) decode(dr *DecodingReader, p unsafe.Pointer) error {
	bytesLen := dr.GetBytesSpan()
	if bytesLen%v.elemSSZ.FixedLen() != 0 {
		return fmt.Errorf("cannot decode basic type array, input has length %d, not compatible with element length %d", bytesLen, v.elemSSZ.FixedLen())
	}

	if endianness.IsLittleEndian || v.elemSSZ.FixedLen() == 1 {
		contentsPtr := v.alloc.MutateLenOrAllocNew(p, bytesLen/v.elemSSZ.FixedLen())
		bytesLimit := v.limit * v.elemSSZ.FixedLen()
		return LittleEndianBasicSeriesDecode(dr, contentsPtr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSlice(v.elemSSZ.Decode, v.elemSSZ.FixedLen(), bytesLen, v.limit, v.alloc, uintptr(v.elemSSZ.FixedLen()), dr, p)
	}
}

func (v *SSZBasicList) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	if dr.IsFuzzMode() {
		return v.decodeFuzzmode(dr, p)
	} else {
		return v.decode(dr, p)
	}
}

func (v *SSZBasicList) DryCheck(dr *DecodingReader) error {
	bytesLen := dr.GetBytesSpan()
	if bytesLen%v.elemSSZ.FixedLen() != 0 {
		return fmt.Errorf("invalid basic type array, input has length %d, not compatible with element length %d", bytesLen, v.elemSSZ.FixedLen())
	}
	bytesLimit := v.limit * v.elemSSZ.FixedLen()
	return BasicSeriesDryCheck(dr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
}

func (v *SSZBasicList) HashTreeRoot(h MerkleFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.ReadSliceHeader(p)

	bytesLen := uint64(sh.Len) * v.elemSSZ.FixedLen()
	bytesLimit := v.limit * v.elemSSZ.FixedLen()
	if endianness.IsLittleEndian || v.elemSSZ.FixedLen() == 1 {
		return h.MixIn(LittleEndianBasicSeriesHTR(h, sh.Data, bytesLen, bytesLimit), uint64(sh.Len))
	} else {
		return h.MixIn(BigEndianBasicSeriesHTR(h, sh.Data, bytesLen, bytesLimit, uint8(v.elemSSZ.FixedLen())), uint64(sh.Len))
	}
}

func (v *SSZBasicList) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)
	length := uint64(sh.Len)
	w.WriteIndent(indent)
	w.Write("[\n")
	w.WriteIndent(indent + 1)
	CallSeries(func(i uint64, p unsafe.Pointer) {
		v.elemSSZ.Pretty(0, w, p)
		if i == length-1 {
			w.Write("\n")
		} else if i%(32/v.elemSSZ.FixedLen()) == 0 {
			w.Write(",\n")
			w.WriteIndent(indent + 1)
		} else {
			w.Write(", ")
		}
	}, length, uintptr(v.elemSSZ.FixedLen()), sh.Data)
	w.WriteIndent(indent)
	w.Write("]")
}
