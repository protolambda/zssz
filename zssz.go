package zssz

import (
	"fmt"
	"io"
	"runtime"
	"unsafe"
	. "zssz/dec"
	. "zssz/enc"
	. "zssz/htr"
	. "zssz/types"
	"zssz/util/ptrutil"
)


func Decode(r io.Reader, bytesLen uint32, val interface{}, sszTyp SSZ) error {
	dr := NewDecodingReader(r)
	dr = dr.Scope(bytesLen)

	p := ptrutil.IfacePtrToPtr(&val)
	err := sszTyp.Decode(dr, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	if err != nil {
		return err
	}
	if readCount := dr.Index(); readCount != bytesLen {
		return fmt.Errorf("read total of %d bytes, but expected %d", readCount, bytesLen)
	}
	return err
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

func HashTreeRoot(h *Hasher, val interface{}, sszTyp SSZ) [32]byte {
	p := ptrutil.IfacePtrToPtr(&val)
	out := sszTyp.HashTreeRoot(h, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
	return out
}

func SigningRoot(h *Hasher, val interface{}, sszTyp SignedSSZ) [32]byte {
	p := unsafe.Pointer(&val)
	return sszTyp.SigningRoot(h, p)
}
