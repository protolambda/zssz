package zssz

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/types"
	"github.com/protolambda/zssz/util/ptrutil"
	"io"
	"reflect"
	"runtime"
)


func Decode(r io.Reader, bytesLen uint32, val interface{}, sszTyp SSZ) error {
	if bytesLen < sszTyp.MinLen() {
		return fmt.Errorf("expected object length is larger than given bytesLen")
	}
	unscoped := NewDecodingReader(r)
	dr, err := unscoped.Scope(bytesLen)
	if err != nil {
		return err
	}
	p := ptrutil.IfacePtrToPtr(&val)
	if err := sszTyp.Decode(dr, p); err != nil {
		return err
	}
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	if readCount := dr.Index(); readCount != bytesLen {
		return fmt.Errorf("read total of %d bytes, but expected %d", readCount, bytesLen)
	}
	return nil
}

// Returns an error if data could not be decoded (something is wrong with the reader).
// Second return value is the amount of bytes that were read, to cut the fuzzing input at if necessary.
func DecodeFuzzBytes(r io.Reader, bytesLen uint32, val interface{}, sszTyp SSZ) (error, uint32) {
	if bytesLen < sszTyp.FuzzReqLen() {
		return fmt.Errorf("expected object fuzzing input length is larger than given bytesLen"), 0
	}
	unscoped := NewDecodingReader(r)
	dr, err := unscoped.Scope(bytesLen)
	if err != nil {
		return err, 0
	}
	dr.EnableFuzzMode()

	p := ptrutil.IfacePtrToPtr(&val)
	if err := sszTyp.Decode(dr, p); err != nil {
		return err, dr.Index()
	}
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&p)
	return nil, dr.Index()
}

func Encode(w io.Writer, val interface{}, sszTyp SSZ) error {
	eb := GetPooledBuffer()
	defer ReleasePooledBuffer(eb)

	p := ptrutil.IfacePtrToPtr(&val)
	sszTyp.Encode(eb, p)

	_, err := eb.WriteTo(w)

	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)

	return err
}

func HashTreeRoot(h HashFn, val interface{}, sszTyp SSZ) [32]byte {
	p := ptrutil.IfacePtrToPtr(&val)
	out := sszTyp.HashTreeRoot(h, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return out
}

func SigningRoot(h HashFn, val interface{}, sszTyp SignedSSZ) [32]byte {
	p := ptrutil.IfacePtrToPtr(&val)
	out := sszTyp.SigningRoot(h, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return out
}

// Gets a SSZ type structure for the given Go type.
// Pass an pointer instance of the type. Can be nil.
// Example: GetSSZ((*MyStruct)(nil))
func GetSSZ(ptr interface{}) SSZ {
	ssz, err := SSZFactory(reflect.TypeOf(ptr).Elem())
	if err != nil {
		panic(err)
	}
	return ssz
}
