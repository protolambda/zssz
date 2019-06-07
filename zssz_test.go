package zssz

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	. "github.com/protolambda/zssz/types"
	"reflect"
	"strings"
	"testing"
)

type emptyTestStruct struct {}

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

type varTestStruct struct {
	A uint16
	B []uint16
	C uint8
}

type complexTestStruct struct {
	A uint16
	B []uint16
	C uint8
	D []byte
	E varTestStruct
	F [4]fixedTestStruct
	G [2]varTestStruct
}

// note: expected strings are in little-endian, hence the seemingly out of order bytes.
var testCases = []struct {
	// name of test
	name string
	// any value
	value interface{}
	// hex formatted, no prefix
	hex string
	// typ getter
	typ reflect.Type
}{
	{"bool F", false, "00", getTyp((*bool)(nil))},
	{"bool T", true, "01", getTyp((*bool)(nil))},
	{"uint8 00", uint8(0x00), "00", getTyp((*uint8)(nil))},
	{"uint8 01", uint8(0x01), "01", getTyp((*uint8)(nil))},
	{"uint8 ab", uint8(0xab), "ab", getTyp((*uint8)(nil))},
	{"uint16 0000", uint16(0x0000), "0000", getTyp((*uint16)(nil))},
	{"uint16 abcd", uint16(0xabcd), "cdab", getTyp((*uint16)(nil))},
	{"uint32 00000000", uint32(0x00000000), "00000000", getTyp((*uint32)(nil))},
	{"uint32 01234567", uint32(0x01234567), "67452301", getTyp((*uint32)(nil))},
	{"small {4567, 0123}", smallTestStruct{0x4567, 0x0123}, "67452301", getTyp((*smallTestStruct)(nil))},
	{"small [4567, 0123]::2", [2]uint16{0x4567, 0x0123}, "67452301", getTyp((*[2]uint16)(nil))},
	{"uint32 01234567", uint32(0x01234567), "67452301", getTyp((*uint32)(nil))},
	{"uint64 0000000000000000", uint64(0x00000000), "0000000000000000", getTyp((*uint64)(nil))},
	{"uint64 0123456789abcdef", uint64(0x0123456789abcdef), "efcdab8967452301", getTyp((*uint64)(nil))},
	{"sig", [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff}, "0100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000ff", getTyp((*[96]byte)(nil))},
	{"emptyTestStruct", emptyTestStruct{}, "", getTyp((*emptyTestStruct)(nil))},
	{"singleFieldTestStruct", singleFieldTestStruct{0xab}, "ab", getTyp((*singleFieldTestStruct)(nil))},
	{"fixedTestStruct", fixedTestStruct{A: 0xab, B: 0xaabbccdd00112233, C: 0x12345678}, "ab33221100ddccbbaa78563412", getTyp((*fixedTestStruct)(nil))},
	{"varTestStruct nil",  varTestStruct{A: 0xabcd, B: nil, C: 0xff}, "cdab07000000ff", getTyp((*varTestStruct)(nil))},
	{"varTestStruct empty", varTestStruct{A: 0xabcd, B: make([]uint16, 0), C: 0xff}, "cdab07000000ff", getTyp((*varTestStruct)(nil))},
	{"varTestStruct some", varTestStruct{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff}, "cdab07000000ff010002000300", getTyp((*varTestStruct)(nil))},
	{"complexTestStruct", complexTestStruct{
		A: 0xaabb,
		B: []uint16{0x1122, 0x3344},
		C: 0xff,
		D: []byte("foobar"),
		E: varTestStruct{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff},
		F: [4]fixedTestStruct{
			{0xcc,0x4242424242424242,0x13371337},
			{0xdd,0x3333333333333333,0xabcdabcd},
			{0xee,0x4444444444444444,0x00112233},
			{0xff,0x5555555555555555,0x44556677}},
		G: [2]varTestStruct{
			{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff},
			{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff}},
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
		"08000000" + "15000000" + // [start G]: local offsets of [2]varTestStruct
		"cdab07000000ff010002000300" +
		"cdab07000000ff010002000300", getTyp((*complexTestStruct)(nil))},
}

func getTyp(ptr interface{}) reflect.Type {
	return reflect.TypeOf(ptr).Elem()
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
			bytesLen := uint32(len(tt.hex)) / 2

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
			t.Logf("root: %x\n", root)
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
			t.Logf("signing root: %x\n", root)
		})
	}
}
