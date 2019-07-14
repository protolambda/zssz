package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	"github.com/protolambda/zssz/util/endianness"
	"github.com/protolambda/zssz/util/ptrutil"
	"reflect"
	"unsafe"
)

type SSZBasicList struct {
	alloc    ptrutil.SliceAllocationFn
	elemKind reflect.Kind
	elemSSZ  *SSZBasic
	limit    uint64
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
		alloc:    ptrutil.MakeSliceAllocFn(typ),
		elemKind: elemKind,
		elemSSZ:  elemSSZ,
		limit:    limit,
	}
	return res, nil
}

func (v *SSZBasicList) FuzzReqLen() uint64 {
	return 4
}

func (v *SSZBasicList) MinLen() uint64 {
	return 0
}

func (v *SSZBasicList) FixedLen() uint64 {
	return 0
}

func (v *SSZBasicList) IsFixed() bool {
	return false
}

func (v *SSZBasicList) Encode(eb *EncodingBuffer, p unsafe.Pointer) {
	sh := ptrutil.ReadSliceHeader(p)

	// we can just write the data as-is in a few contexts:
	// - if we're in a little endian architecture
	// - if there is no endianness to deal with
	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		LittleEndianBasicSeriesEncode(eb, sh.Data, uint64(sh.Len)*v.elemSSZ.Length)
	} else {
		EncodeFixedSeries(v.elemSSZ.Encoder, uint64(sh.Len), uintptr(v.elemSSZ.Length), eb, sh.Data)
	}
}

func (v *SSZBasicList) decodeFuzzmode(dr *DecodingReader, p unsafe.Pointer) error {
	x, err := dr.ReadUint64()
	if err != nil {
		return err
	}
	span := dr.GetBytesSpan()
	if byteLimit := v.limit * v.elemSSZ.FuzzReqLen(); byteLimit > span {
		span = byteLimit
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
