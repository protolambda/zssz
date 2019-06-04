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
	elemKind reflect.Kind
	elemSSZ  *SSZBasic
}

func NewSSZBasicList(typ reflect.Type) (*SSZBasicList, error) {
	if typ.Kind() != reflect.Slice {
		return nil, fmt.Errorf("typ is not a dynamic-length array")
	}
	elemTyp := typ.Elem()
	elemKind := elemTyp.Kind()
	elemSSZ, err := GetBasicSSZElemType(elemKind)
	if err != nil {
		return nil, err
	}
	if elemSSZ.Length != uint32(elemTyp.Size()) {
		return nil, fmt.Errorf("basic element type has different size than SSZ type unexpectedly, ssz: %d, go: %d", elemSSZ.Length, elemTyp.Size())
	}

	res := &SSZBasicList{
		elemKind: elemKind,
		elemSSZ:  elemSSZ,
	}
	return res, nil
}

func (v *SSZBasicList) FuzzReqLen() uint32 {
	return 4
}

func (v *SSZBasicList) MinLen() uint32 {
	return 0
}

func (v *SSZBasicList) FixedLen() uint32 {
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
		LittleEndianBasicSeriesEncode(eb, sh.Data, uint32(sh.Len)*v.elemSSZ.Length)
	} else {
		EncodeFixedSeries(v.elemSSZ.Encoder, uint32(sh.Len), uintptr(v.elemSSZ.Length), eb, sh.Data)
	}
}

func (v *SSZBasicList) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	var bytesLen uint32
	if dr.IsFuzzMode() {
		span, err := dr.ReadUint32()
		if err != nil {
			return err
		}
		bytesLen = span % dr.GetBytesSpan()
		bytesLen = bytesLen % v.elemSSZ.Length
	} else {
		bytesLen = dr.Max() - dr.Index()
		if bytesLen%v.elemSSZ.Length != 0 {
			return fmt.Errorf("cannot decode basic type array, input has is")
		}
	}
	elemMemSize := uintptr(v.elemSSZ.Length)

	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		contentsPtr := ptrutil.AllocateSliceSpaceAndBind(p, bytesLen/v.elemSSZ.Length, elemMemSize)
		return LittleEndianBasicSeriesDecode(dr, contentsPtr, bytesLen, v.elemKind == reflect.Bool)
	} else {
		return DecodeFixedSlice(v.elemSSZ.Decoder, v.elemSSZ.FixedLen(), bytesLen, elemMemSize, dr, p)
	}
}

func (v *SSZBasicList) HashTreeRoot(h *Hasher, p unsafe.Pointer) [32]byte {
	//elemSize := v.elemMemSize
	sh := ptrutil.ReadSliceHeader(p)

	if endianness.IsLittleEndian || v.elemSSZ.Length == 1 {
		return LittleEndianBasicSeriesHTR(h, sh.Data, uint32(sh.Len), uint32(sh.Len)*v.elemSSZ.Length, v.elemSSZ.ChunkPow)
	} else {
		return BigEndianBasicSeriesHTR(h, sh.Data, uint32(sh.Len), uint32(sh.Len)*v.elemSSZ.Length, v.elemSSZ.ChunkPow)
	}
}
