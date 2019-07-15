package zssz

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/protolambda/zssz/bitfields"
	. "github.com/protolambda/zssz/types"
	"reflect"
	"strings"
	"testing"
)

type emptyTestStruct struct{}

type singleFieldTestStruct struct {
	A byte
}

type smallTestStruct struct {
	A uint16
	B uint16
}

type fixedTestStruct struct {
	A uint8
	B uint64
	C uint32
}

type uint16List128 []uint16

func (li *uint16List128) Limit() uint64 {
	return 128
}

type uint16List1024 []uint16

func (li *uint16List1024) Limit() uint64 {
	return 1024
}

type bytelist256 []byte

func (li *bytelist256) Limit() uint64 {
	return 256
}

type VarTestStruct struct {
	A uint16
	B uint16List1024
	C uint8
}

type complexTestStruct struct {
	A uint16
	B uint16List128
	C uint8
	D bytelist256
	E VarTestStruct
	F [4]fixedTestStruct
	G [2]VarTestStruct
}

type embeddingStruct struct {
	A VarTestStruct
	VarTestStruct // squash field by embedding (must be a public type)
	B   uint16
	Foo smallTestStruct `ssz:"squash"` // Squash field explicitly
}

type Squash1 struct {
	A uint8
	D *uint32 `ssz:"omit"`
	B uint64
	C uint32
}

type Squash2 struct {
	D uint32
	E uint8 `ssz:"omit"`
	Squash1
	More Squash1 `ssz:"squash"`
}

type Squash3 struct {
	Foo Squash1 `ssz:"squash"`
	Squash1
	X   Squash2 `ssz:"squash"`
	Bar Squash1 `ssz:"squash"`
	Squash2
}

func chunk(v string) string {
	res := [32]byte{}
	data, _ := hex.DecodeString(v)
	copy(res[:], data)
	return hex.EncodeToString(res[:])
}

func h(a string, b string) string {
	aBytes, _ := hex.DecodeString(a)
	bBytes, _ := hex.DecodeString(b)
	data := append(append(make([]byte, 0, 64), aBytes...), bBytes...)
	res := sha256.Sum256(data)
	return hex.EncodeToString(res[:])
}

func merge(a string, branch []string) (out string) {
	out = a
	for _, b := range branch {
		out = h(out, b)
	}
	return
}

func repeat(v string, count int) (out string) {
	for i := 0; i < count; i++ {
		out += v
	}
	return
}

// Many different Bitvector types for testing

type bitvec513 [64 + 1]byte

func (_ *bitvec513) BitLen() uint64 { return 513 }

type bitvec512 [64]byte

func (_ *bitvec512) BitLen() uint64 { return 512 }

type bitvec16 [2]byte

func (_ *bitvec16) BitLen() uint64 { return 16 }

type bitvec10 [2]byte

func (_ *bitvec10) BitLen() uint64 { return 10 }

type bitvec8 [1]byte

func (_ *bitvec8) BitLen() uint64 { return 8 }

type bitvec4 [1]byte

func (_ *bitvec4) BitLen() uint64 { return 4 }

type bitvec3 [1]byte

func (_ *bitvec3) BitLen() uint64 { return 3 }

// Many different Bitlist types for testing

type bitlist513 []byte

func (_ *bitlist513) Limit() uint64 { return 513 }
func (b bitlist513) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist512 []byte

func (_ *bitlist512) Limit() uint64 { return 512 }
func (b bitlist512) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist16 []byte

func (_ *bitlist16) Limit() uint64 { return 16 }
func (b bitlist16) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist10 []byte

func (_ *bitlist10) Limit() uint64 { return 10 }
func (b bitlist10) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist8 []byte

func (_ *bitlist8) Limit() uint64 { return 8 }
func (b bitlist8) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist4 []byte

func (_ *bitlist4) Limit() uint64 { return 4 }
func (b bitlist4) BitLen() uint64 { return bitfields.BitlistLen(b) }

type bitlist3 []byte

func (_ *bitlist3) Limit() uint64 { return 3 }
func (b bitlist3) BitLen() uint64 { return bitfields.BitlistLen(b) }

// Some list types for testing

type list32uint16 []uint16

func (_ *list32uint16) Limit() uint64 { return 32 }

type list128uint32 []uint32

func (_ *list128uint32) Limit() uint64 { return 128 }

type list64bytes32 [][32]byte

func (_ *list64bytes32) Limit() uint64 { return 64 }

type list128bytes32 [][32]byte

func (_ *list128bytes32) Limit() uint64 { return 128 }

func getTyp(ptr interface{}) reflect.Type {
	return reflect.TypeOf(ptr).Elem()
}

type sszTestCase struct {
	// name of test
	name string
	// any value
	value interface{}
	// hex formatted, no prefix
	hex string
	// hex formatted, no prefix
	root string
	// typ getter
	typ reflect.Type
}

// note: expected strings are in little-endian, hence the seemingly out of order bytes.
var testCases []sszTestCase

func init() {
	var zeroHashes = []string{chunk("")}

	for layer := 1; layer < 32; layer++ {
		zeroHashes = append(zeroHashes, h(zeroHashes[layer-1], zeroHashes[layer-1]))
	}

	testCases = []sszTestCase{
		{"bool F", false, "00", chunk("00"), getTyp((*bool)(nil))},
		{"bool T", true, "01", chunk("01"), getTyp((*bool)(nil))},
		{"bitvector TTFTFTFF", bitvec8{0x2b}, "2b", chunk("2b"), getTyp((*bitvec8)(nil))},
		{"bitlist TTFTFTFF", bitlist8{0x2b, 0x01}, "2b01", h(chunk("2b"), chunk("08")), getTyp((*bitlist8)(nil))},
		{"bitvector FTFT", bitvec4{0x0a}, "0a", chunk("0a"), getTyp((*bitvec4)(nil))},
		{"bitlist FTFT", bitlist4{0x1a}, "1a", h(chunk("0a"), chunk("04")), getTyp((*bitlist4)(nil))},
		{"bitvector FTF", bitvec3{0x02}, "02", chunk("02"), getTyp((*bitvec3)(nil))},
		{"bitlist FTF", bitlist3{0x0a}, "0a", h(chunk("02"), chunk("03")), getTyp((*bitlist3)(nil))},
		{"bitvector TFTFFFTTFT", bitvec10{0xc5, 0x02}, "c502", chunk("c502"), getTyp((*bitvec10)(nil))},
		{"bitlist TFTFFFTTFT", bitlist10{0xc5, 0x06}, "c506", h(chunk("c502"), chunk("0A")), getTyp((*bitlist10)(nil))},
		{"bitvector TFTFFFTTFTFFFFTT", bitvec16{0xc5, 0xc2}, "c5c2", chunk("c5c2"), getTyp((*bitvec16)(nil))},
		{"bitlist TFTFFFTTFTFFFFTT", bitlist16{0xc5, 0xc2, 0x01}, "c5c201", h(chunk("c5c2"), chunk("10")), getTyp((*bitlist16)(nil))},
		{"long bitvector", bitvec512{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		}, repeat("ff", 64), h(repeat("ff", 32), repeat("ff", 32)), getTyp((*bitvec512)(nil)),
		},
		{"long bitlist", bitlist512{7}, "03", h(h(chunk("03"), chunk("")), chunk("02")), getTyp((*bitlist512)(nil))},
		{"long bitlist filled", bitlist512{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x01,
		}, repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), chunk("0002")), getTyp((*bitlist512)(nil))},
		{"odd bitvector filled", bitvec513{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x01,
		}, repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk(""))), getTyp((*bitvec513)(nil))},
		{"odd bitlist filled", bitlist513{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x03,
		}, repeat("ff", 64) + "03", h(h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk(""))), chunk("0102")), getTyp((*bitlist513)(nil))},
		{"uint8 00", uint8(0x00), "00", chunk("00"), getTyp((*uint8)(nil))},
		{"uint8 01", uint8(0x01), "01", chunk("01"), getTyp((*uint8)(nil))},
		{"uint8 ab", uint8(0xab), "ab", chunk("ab"), getTyp((*uint8)(nil))},
		{"uint16 0000", uint16(0x0000), "0000", chunk("0000"), getTyp((*uint16)(nil))},
		{"uint16 abcd", uint16(0xabcd), "cdab", chunk("cdab"), getTyp((*uint16)(nil))},
		{"uint32 00000000", uint32(0x00000000), "00000000", chunk("00000000"), getTyp((*uint32)(nil))},
		{"uint32 01234567", uint32(0x01234567), "67452301", chunk("67452301"), getTyp((*uint32)(nil))},
		{"small {4567, 0123}", smallTestStruct{0x4567, 0x0123}, "67452301", h(chunk("6745"), chunk("2301")), getTyp((*smallTestStruct)(nil))},
		{"small [4567, 0123]::2", [2]uint16{0x4567, 0x0123}, "67452301", chunk("67452301"), getTyp((*[2]uint16)(nil))},
		{"uint32 01234567", uint32(0x01234567), "67452301", chunk("67452301"), getTyp((*uint32)(nil))},
		{"uint64 0000000000000000", uint64(0x00000000), "0000000000000000", chunk("0000000000000000"), getTyp((*uint64)(nil))},
		{"uint64 0123456789abcdef", uint64(0x0123456789abcdef), "efcdab8967452301", chunk("efcdab8967452301"), getTyp((*uint64)(nil))},
		{"sig", [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff},
			"01" + repeat("00", 31) + "02" + repeat("00", 31) + "03" + repeat("00", 30) + "ff",
			h(h(chunk("01"), chunk("02")), h("03"+repeat("00", 30)+"ff", chunk(""))), getTyp((*[96]byte)(nil))},
		{"emptyTestStruct", emptyTestStruct{}, "", chunk(""), getTyp((*emptyTestStruct)(nil))},
		{"singleFieldTestStruct", singleFieldTestStruct{0xab}, "ab", chunk("ab"), getTyp((*singleFieldTestStruct)(nil))},

		{"uint16 list", list32uint16{0xaabb, 0xc0ad, 0xeeff}, "bbaaadc0ffee",
			h(h(chunk("bbaaadc0ffee"), chunk("")), chunk("03000000")), // max length: 32 * 2 = 64 bytes = 2 chunks
			getTyp((*list32uint16)(nil)),
		},
		{"uint32 list", list128uint32{0xaabb, 0xc0ad, 0xeeff}, "bbaa0000adc00000ffee0000",
			// max length: 128 * 4 = 512 bytes = 16 chunks
			h(merge(chunk("bbaa0000adc00000ffee0000"), zeroHashes[0:4]), chunk("03000000")),
			getTyp((*list128uint32)(nil)),
		},
		{"bytes32 list", list64bytes32{[32]byte{0xbb, 0xaa}, [32]byte{0xad, 0xc0}, [32]byte{0xff, 0xee}},
			"bbaa000000000000000000000000000000000000000000000000000000000000" +
				"adc0000000000000000000000000000000000000000000000000000000000000" +
				"ffee000000000000000000000000000000000000000000000000000000000000",
			h(merge(h(h(chunk("bbaa"), chunk("adc0")), h(chunk("ffee"), chunk(""))), zeroHashes[2:6]), chunk("03000000")),
			getTyp((*list64bytes32)(nil)),
		},
		{"bytes32 list long", list128bytes32{
			{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10},
			{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19},
		},
			"01" + repeat("00", 31) + "02" + repeat("00", 31) +
				"03" + repeat("00", 31) + "04" + repeat("00", 31) +
				"05" + repeat("00", 31) + "06" + repeat("00", 31) +
				"07" + repeat("00", 31) + "08" + repeat("00", 31) +
				"09" + repeat("00", 31) + "0a" + repeat("00", 31) +
				"0b" + repeat("00", 31) + "0c" + repeat("00", 31) +
				"0d" + repeat("00", 31) + "0e" + repeat("00", 31) +
				"0f" + repeat("00", 31) + "10" + repeat("00", 31) +
				"11" + repeat("00", 31) + "12" + repeat("00", 31) +
				"13" + repeat("00", 31),
			h(merge(
				h(
					h(
						h(
							h(h(chunk("01"), chunk("02")), h(chunk("03"), chunk("04"))),
							h(h(chunk("05"), chunk("06")), h(chunk("07"), chunk("08"))),
						),
						h(
							h(h(chunk("09"), chunk("0a")), h(chunk("0b"), chunk("0c"))),
							h(h(chunk("0d"), chunk("0e")), h(chunk("0f"), chunk("10"))),
						),
					),
					h(
						h(
							h(h(chunk("11"), chunk("12")), h(chunk("13"), chunk(""))),
							zeroHashes[2],
						),
						zeroHashes[3],
					),
				),
				// 128 chunks = 7 deep
				zeroHashes[5:7]), chunk("13000000")),
			getTyp((*list128bytes32)(nil)),
		},
		{"fixedTestStruct", fixedTestStruct{A: 0xab, B: 0xaabbccdd00112233, C: 0x12345678}, "ab33221100ddccbbaa78563412",
			h(h(chunk("ab"), chunk("33221100ddccbbaa")), h(chunk("78563412"), chunk(""))), getTyp((*fixedTestStruct)(nil))},
		{"VarTestStruct nil", VarTestStruct{A: 0xabcd, B: nil, C: 0xff}, "cdab07000000ff",
			// log2(1024*2/32)= 6 deep
			h(h(chunk("cdab"), h(zeroHashes[6], chunk("00000000"))), h(chunk("ff"), chunk(""))), getTyp((*VarTestStruct)(nil))},
		{"VarTestStruct empty", VarTestStruct{A: 0xabcd, B: make([]uint16, 0), C: 0xff}, "cdab07000000ff",
			h(h(chunk("cdab"), h(zeroHashes[6], chunk("00000000"))), h(chunk("ff"), chunk(""))), getTyp((*VarTestStruct)(nil))},
		{"VarTestStruct some", VarTestStruct{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff}, "cdab07000000ff010002000300",
			h(
				h(
					chunk("cdab"),
					h(
						merge(
							chunk("010002000300"),
							zeroHashes[0:6],
						),
						chunk("03000000"), // length mix in
					),
				),
				h(chunk("ff"), chunk("")),
			),
			getTyp((*VarTestStruct)(nil))},
		{"complexTestStruct",
			complexTestStruct{
				A: 0xaabb,
				B: uint16List128{0x1122, 0x3344},
				C: 0xff,
				D: bytelist256("foobar"),
				E: VarTestStruct{A: 0xabcd, B: uint16List1024{1, 2, 3}, C: 0xff},
				F: [4]fixedTestStruct{
					{0xcc, 0x4242424242424242, 0x13371337},
					{0xdd, 0x3333333333333333, 0xabcdabcd},
					{0xee, 0x4444444444444444, 0x00112233},
					{0xff, 0x5555555555555555, 0x44556677}},
				G: [2]VarTestStruct{
					{A: 0xdead, B: []uint16{1, 2, 3}, C: 0x11},
					{A: 0xbeef, B: []uint16{4, 5, 6}, C: 0x22}},
			},
			"bbaa" +
				"47000000" + // offset of B, []uint16
				"ff" +
				"4b000000" + // offset of foobar
				"51000000" + // offset of E
				"cc424242424242424237133713" +
				"dd3333333333333333cdabcdab" +
				"ee444444444444444433221100" +
				"ff555555555555555577665544" +
				"5e000000" + // pointer to G
				"22114433" + // contents of B
				"666f6f626172" + // foobar
				"cdab07000000ff010002000300" + // contents of E
				"08000000" + "15000000" + // [start G]: local offsets of [2]VarTestStruct
				"adde0700000011010002000300" +
				"efbe0700000022040005000600",
			h(
				h(
					h( // A and B
						chunk("bbaa"),
						h(merge(chunk("22114433"), zeroHashes[0:3]), chunk("02000000")), // 2*128/32 = 8 chunks
					),
					h( // C and D
						chunk("ff"),
						h(merge(chunk("666f6f626172"), zeroHashes[0:3]), chunk("06000000")), // 256/32 = 8 chunks
					),
				),
				h(
					h( // E and F
						h(h(chunk("cdab"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
							h(chunk("ff"), chunk(""))),
						h(
							h(
								h(h(chunk("cc"), chunk("4242424242424242")), h(chunk("37133713"), chunk(""))),
								h(h(chunk("dd"), chunk("3333333333333333")), h(chunk("cdabcdab"), chunk(""))),
							),
							h(
								h(h(chunk("ee"), chunk("4444444444444444")), h(chunk("33221100"), chunk(""))),
								h(h(chunk("ff"), chunk("5555555555555555")), h(chunk("77665544"), chunk(""))),
							),
						),
					),
					h( // G and padding
						h(
							h(h(chunk("adde"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
								h(chunk("11"), chunk(""))),
							h(h(chunk("efbe"), h(merge(chunk("040005000600"), zeroHashes[0:6]), chunk("03000000"))),
								h(chunk("22"), chunk(""))),
						),
						chunk(""),
					),
				),
			),
			getTyp((*complexTestStruct)(nil))},
		{"embeddingStruct",
			embeddingStruct{
				A:             VarTestStruct{A: 0xabcd, B: uint16List1024{1, 2, 3}, C: 0xff},
				VarTestStruct: VarTestStruct{A: 0xeeff, B: uint16List1024{4, 5, 6, 7, 8}, C: 0xaa},
				B:             0x1234,
				Foo: smallTestStruct{
					A: 0x4567,
					B: 0x8901,
				},
			},
			"11000000" + // offset to dynamic part A
				"ffee" + "1e000000" + "aa" + // embedded VarTestStruct, fixed part
				"3412" + // B
				"6745" + "0189" + // embedded smallTestStruct
				"cdab07000000ff010002000300" + // A, contents
				"04000500060007000800", // embedded VarTestStruct, dynamic part
			h(
				h(
					h(
						// A
						h(
							h(chunk("cdab"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
							h(chunk("ff"), chunk("")),
						),
						// embedded
						chunk("ffee"),
					),
					h(
						// embedded continued
						h(merge(chunk("04000500060007000800"), zeroHashes[0:6]), chunk("05000000")),
						chunk("aa"),
					),
				),
				h(
					h(
						chunk("3412"), // B
						chunk("6745"), // embedded smallTestStruct
					),
					h(
						chunk("0189"), // embedded continued
						chunk(""),
					),
				),
			),
			getTyp((*embeddingStruct)(nil))},
		{"squash chaos", Squash3{
			Foo:     Squash1{01, nil, 0xa8a7a6a5a4a3a2a1, 0xaabbccdd},
			Squash1: Squash1{02, nil, 0xb8b7b6b5b4b3b2b1, 0x00001111},
			X: Squash2{
				D:       0x11223344,
				E:       0, // omitted
				Squash1: Squash1{03, nil, 0xc8c7c6c5c4c3c2c1, 0x22223333},
				More:    Squash1{04, nil, 0xd8d7d6d5d4d3d2d1, 0x42424242},
			},
			Bar: Squash1{0xab, nil, 0x1000000000000001, 0x12341234},
			Squash2: Squash2{
				D:       0x12345678,
				E:       0, // omitted
				Squash1: Squash1{05, nil, 0xe8e7e6e5e4e3e2e1, 0x55665566},
				More:    Squash1{06, nil, 0xf8f7f6f5f4f3f2f1, 0x78787878},
			},
		}, "01" + "a1a2a3a4a5a6a7a8" + "ddccbbaa" +
			"02" + "b1b2b3b4b5b6b7b8" + "11110000" +
			"44332211" +
			"03" + "c1c2c3c4c5c6c7c8" + "33332222" +
			"04" + "d1d2d3d4d5d6d7d8" + "42424242" +
			"ab" + "0100000000000010" + "34123412" +
			"78563412" +
			"05" + "e1e2e3e4e5e6e7e8" + "66556655" +
			"06" + "f1f2f3f4f5f6f7f8" + "78787878",
			h(
				h(
					h(
						h(h(chunk("01"), chunk("a1a2a3a4a5a6a7a8")), h(chunk("ddccbbaa"), chunk("02"))),
						h(h(chunk("b1b2b3b4b5b6b7b8"), chunk("11110000")), h(chunk("44332211"), chunk("03"))),
					),
					h(
						h(h(chunk("c1c2c3c4c5c6c7c8"), chunk("33332222")), h(chunk("04"), chunk("d1d2d3d4d5d6d7d8"))),
						h(h(chunk("42424242"), chunk("ab")), h(chunk("0100000000000010"), chunk("34123412"))),
					),
				),
				h(
					h(
						h(h(chunk("78563412"), chunk("05")), h(chunk("e1e2e3e4e5e6e7e8"), chunk("66556655"))),
						h(h(chunk("06"), chunk("f1f2f3f4f5f6f7f8")), h(chunk("78787878"), chunk(""))),
					),
					zeroHashes[3],
				),
			),
			getTyp((*Squash3)(nil))},
	}
}

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			if err := Encode(bufWriter, tt.value, sszTyp); err != nil {
				t.Fatal(err)
			}
			if err := bufWriter.Flush(); err != nil {
				t.Fatal(err)
			}
			data := buf.Bytes()
			if res := fmt.Sprintf("%x", data); res != tt.hex {
				t.Fatalf("encoded different data:\n     got %s\nexpected %s", res, tt.hex)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(data)
			// For dynamic types, we need to pass the length of the message to the decoder.
			// See SSZ-envelope discussion
			bytesLen := uint64(len(tt.hex)) / 2

			destination := reflect.New(tt.typ).Interface()
			if err := Decode(r, bytesLen, destination, sszTyp); err != nil {
				t.Fatal(err)
			}
			res, err := json.Marshal(destination)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			// adjust expected json string. No effective difference between null and an empty slice. We prefer nil.
			if adjusted := strings.ReplaceAll(string(expected), "[]", "null"); string(res) != adjusted {
				t.Fatalf("decoded different data:\n     got %s\nexpected %s", res, adjusted)
			}
		})
	}
}

func TestHashTreeRoot(t *testing.T) {
	var buf bytes.Buffer

	// re-use a hash function
	sha := sha256.New()
	hashFn := func(input []byte) (out [32]byte) {
		sha.Reset()
		sha.Write(input)
		copy(out[:], sha.Sum(nil))
		return
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			root := HashTreeRoot(hashFn, tt.value, sszTyp)
			res := hex.EncodeToString(root[:])
			if res != tt.root {
				t.Errorf("Expected root %s but got %s", tt.root, res)
			}
		})
	}
}

func TestSigningRoot(t *testing.T) {
	var buf bytes.Buffer

	// re-use a hash function
	sha := sha256.New()
	hashFn := func(input []byte) (out [32]byte) {
		sha.Reset()
		sha.Write(input)
		copy(out[:], sha.Sum(nil))
		return
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			signedSSZ, ok := sszTyp.(SignedSSZ)
			if !ok {
				// only test signing root for applicable types
				return
			}
			root := SigningRoot(hashFn, tt.value, signedSSZ)
			res := hex.EncodeToString(root[:])
			if res == tt.root && root != ([32]byte{}) {
				t.Errorf("Signing root is not different than hash-tree-root. "+
					"Expected root: %s but got %s (should be different)", tt.root, res)
			}
			t.Logf("signing root: %x\n", root)
		})
	}
}
