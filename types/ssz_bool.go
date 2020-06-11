package types

import (
	"fmt"
	. "github.com/protolambda/zssz/dec"
	. "github.com/protolambda/zssz/enc"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/pretty"
	"unsafe"
)

type SSZBool struct{}

func (v SSZBool) FuzzMinLen() uint64 {
	return 1
}

func (v SSZBool) FuzzMaxLen() uint64 {
	return 1
}

func (v SSZBool) MinLen() uint64 {
	return 1
}

func (v SSZBool) MaxLen() uint64 {
	return 1
}

func (v SSZBool) FixedLen() uint64 {
	return 1
}

func (v SSZBool) IsFixed() bool {
	return true
}

func (v SSZBool) SizeOf(p unsafe.Pointer) uint64 {
	return 1
}

func (v SSZBool) Encode(eb *EncodingWriter, p unsafe.Pointer) error {
	return eb.WriteByte(*(*byte)(p))
}

func (v SSZBool) Decode(dr *DecodingReader, p unsafe.Pointer) error {
	b, err := dr.ReadByte()
	if err != nil {
		return err
	}
	if b == 0x00 {
		*(*bool)(p) = false
		return nil
	} else if b == 0x01 {
		*(*bool)(p) = true
		return nil
	} else {
		if dr.IsFuzzMode() {
			// just make a valid random bool
			*(*bool)(p) = b&1 != 0
			return nil
		} else {
			return fmt.Errorf("bool value is invalid")
		}
	}
}

func (v SSZBool) DryCheck(dr *DecodingReader) error {
	b, err := dr.ReadByte()
	if err != nil {
		return err
	}
	if b > 1 {
		return fmt.Errorf("bool value is invalid")
	}
	return nil
}

func (v SSZBool) HashTreeRoot(h MerkleFn, p unsafe.Pointer) (out [32]byte) {
	out[0] = *(*byte)(p)
	return
}

func (v SSZBool) Pretty(indent uint32, w *PrettyWriter, p unsafe.Pointer) {
	w.WriteIndent(indent)
	if *(*bool)(p) {
		w.Write("True")
	} else {
		w.Write("False")
	}
}
