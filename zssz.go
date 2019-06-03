package zssz

import (
	"io"
	"runtime"
	"unsafe"
	. "zssz/dec"
	. "zssz/enc"
	. "zssz/htr"
	. "zssz/types"
	"zssz/util/ptrutil"
)


func Decode(r io.Reader, val interface{}, sszTyp SSZ) error {
	dr := NewDecodingReader(r)

	p := ptrutil.IfacePtrToPtr(&val)
	err := sszTyp.Decode(dr, p)
	// make sure the data of the object is kept around up to this point.
	runtime.KeepAlive(&val)
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
