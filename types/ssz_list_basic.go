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
	elemSSZ   *SSZBasic
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
	if elemSSZ.Length != uint64(elemTyp.Size()) {
		return nil, fmt.Errorf("basic element type has different size than SSZ type unexpectedly, ssz: %d, go: %d", elemSSZ.Length, elemTyp.Size())
	}

	res := &SSZBasicList{
		alloc:     ptrutil.MakeSliceAllocFn(typ),
		elemKind:  elemKind,
		elemSSZ:   elemSSZ,
		limit:     limit,
		byteLimit: limit * elemSSZ.Length,
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
	return uint64(sh.Len) * v.elemSSZ.Length
}

func (v *SSZBasicList) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	sh := ptrutil.ReadSliceHeader(p)

	// we can just write the data as-is in a few contexts:
	// - if we're in a little endian architecture
	// - if there is no endianness to deal with
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesEncode(eb, sh.Data, uint64(sh.Len)*v.elemSSZ.Length)
	} else {
		return EncodeFixedSeries(v.elemSSZ.Encoder, uint64(sh.Len), uintptr(v.elemSSZ.Length), eb, sh.Data)
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
	bytesLen -= bytesLen % v.elemSSZ.Length

	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		contentsPtr := v.alloc(p, bytesLen/v.elemSSZ.Length)
		bytesLimit := v.limit * v.elemSSZ.Length
		return LittleEndianBasicSeriesDecode(dr, contentsPtr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSlice(v.elemSSZ.Decoder, v.elemSSZ.Length, bytesLen, v.limit, v.alloc, uintptr(v.elemSSZ.Length), dr, p)
	}
}

func (v *SSZBasicList) decode(dr *DecodingReader, p unsafe.Pointer) error {
	bytesLen := dr.GetBytesSpan()
	if bytesLen%v.elemSSZ.Length != 0 {
		return fmt.Errorf("cannot decode basic type array, input has length %d, not compatible with element length %d", bytesLen, v.elemSSZ.Length)
	}

	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		contentsPtr := v.alloc(p, bytesLen/v.elemSSZ.Length)
		bytesLimit := v.limit * v.elemSSZ.Length
		return LittleEndianBasicSeriesDecode(dr, contentsPtr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSlice(v.elemSSZ.Decoder, v.elemSSZ.Length, bytesLen, v.limit, v.alloc, uintptr(v.elemSSZ.Length), dr, p)
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
	if bytesLen%v.elemSSZ.Length != 0 {
		return fmt.Errorf("invalid basic type array, input has length %d, not compatible with element length %d", bytesLen, v.elemSSZ.Length)
	}
	bytesLimit := v.limit * v.elemSSZ.Length
	return BasicSeriesDryCheck(dr, bytesLen, bytesLimit, v.elemKind == reflect.Bool)
}

func (v *SSZBasicList) HashTreeRoot(h HashFn, p unsafe.Pointer) [32]byte {
	sh := ptrutil.ReadSliceHeader(p)

	bytesLen := uint64(sh.Len) * v.elemSSZ.Length
	bytesLimit := v.limit * v.elemSSZ.Length
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return h.MixIn(LittleEndianBasicSeriesHTR(h, sh.Data, bytesLen, bytesLimit), uint64(sh.Len))
	} else {
		return h.MixIn(BigEndianBasicSeriesHTR(h, sh.Data, bytesLen, bytesLimit, uint8(v.elemSSZ.Length)), uint64(sh.Len))
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
		} else if i%(32/v.elemSSZ.Length) == 0 {
			w.Write(",\n")
			w.WriteIndent(indent + 1)
		} else {
			w.Write(", ")
		}
	}, length, uintptr(v.elemSSZ.Length), sh.Data)
	w.WriteIndent(indent)
	w.Write("]")
}
