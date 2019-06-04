package zssz

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz/types"
	"reflect"
	"strings"
	"testing"
)

type getTypFn func() reflect.Type

func booltest() reflect.Type {
	return reflect.TypeOf(true)
}
func uint8test() reflect.Type {
	return reflect.TypeOf(uint8(0))
}
func uint16test() reflect.Type {
	return reflect.TypeOf(uint16(0))
}
func uint32test() reflect.Type {
	return reflect.TypeOf(uint32(0))
}
func uint64test() reflect.Type {
	return reflect.TypeOf(uint64(0))
}

type fixedTestStruct struct {
	A uint8
	B uint64
	C uint32
}
func fixedTestStructTest() reflect.Type {
	return reflect.TypeOf(new(fixedTestStruct)).Elem()
}

type varTestStruct struct {
	A uint16
	B []uint16
	C uint8
}

func varTestStructTest() reflect.Type {
	return reflect.TypeOf(new(varTestStruct)).Elem()
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

func complexTestStructTest() reflect.Type {
	return reflect.TypeOf(new(complexTestStruct)).Elem()
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
	getTyp getTypFn
}{
	{"bool F", false, "00", booltest},
	{"bool T", true, "01", booltest},
	{"uint8 00", uint8(0x00), "00", uint8test},
	{"uint8 ab", uint8(0xab), "ab", uint8test},
	{"uint16 0000", uint16(0x0000), "0000", uint16test},
	{"uint16 abcd", uint16(0xabcd), "cdab", uint16test},
	{"uint32 00000000", uint32(0x00000000), "00000000", uint32test},
	{"uint32 01234567", uint32(0x01234567), "67452301", uint32test},
	{"uint64 0000000000000000", uint64(0x00000000), "0000000000000000", uint64test},
	{"uint64 0123456789abcdef", uint64(0x0123456789abcdef), "efcdab8967452301", uint64test},
	{"fixedTestStruct", &fixedTestStruct{A: 0xab, B: 0xaabbccdd00112233, C: 0x12345678}, "ab33221100ddccbbaa78563412", fixedTestStructTest},
	{"varTestStruct nil",  varTestStruct{A: 0xabcd, B: nil, C: 0xff}, "cdab07000000ff", varTestStructTest},
	{"varTestStruct empty", varTestStruct{A: 0xabcd, B: make([]uint16, 0), C: 0xff}, "cdab07000000ff", varTestStructTest},
	{"varTestStruct some", varTestStruct{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff}, "cdab07000000ff010002000300", varTestStructTest},
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
		"cdab07000000ff010002000300", complexTestStructTest},
}

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			typ := tt.getTyp()
			sszTyp, err := SSZFactory(typ)
			if err != nil {
				t.Error(err)
			}
			if err := Encode(bufWriter, tt.value, sszTyp); err != nil {
				t.Error(err)
			}
			if err := bufWriter.Flush(); err != nil {
				t.Error(err)
			}
			data := buf.Bytes()
			if res := fmt.Sprintf("%x", data); res != tt.hex {
				t.Errorf("encoded different data:\n     got %s\nexpected %s", res, tt.hex)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	var buf bytes.Buffer

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			typ := tt.getTyp()
			sszTyp, err := SSZFactory(typ)
			if err != nil {
				t.Error(err)
			}
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Error(err)
			}
			buf.Write(data)
			// For dynamic types, we need to pass the length of the message to the decoder.
			// See SSZ-envelope discussion
			bytesLen := uint32(len(tt.hex)) / 2

			bufReader := bufio.NewReader(&buf)
			destination := reflect.New(typ).Interface()
			if err := Decode(bufReader, bytesLen, destination, sszTyp); err != nil {
				t.Error(err)
			}
			res, err := json.Marshal(destination)
			if err != nil {
				t.Error(err)
			}
			expected, err := json.Marshal(tt.value)
			if err != nil {
				t.Error(err)
			}
			// adjust expected json string. No difference between null and an empty slice.
			if adjusted := strings.ReplaceAll(string(expected), "null", "[]"); string(res) != adjusted {
				t.Errorf("decoded different data:\n     got %s\nexpected %s", res, adjusted)
			}
		})
	}
}

func TestHashTreeRoot(t *testing.T) {
	var buf bytes.Buffer

	sha := sha256.New()
	hashFn := func(input []byte) (out [32]byte) {
		sha.Reset()
		sha.Write(input)
		copy(out[:], sha.Sum(nil))
		return
	}
	// re-use a hash function, and change it if you like
	hasher := htr.NewHasher(hashFn)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			typ := tt.getTyp()
			sszTyp, err := SSZFactory(typ)
			if err != nil {
				t.Error(err)
			}
			root := HashTreeRoot(hasher, tt.value, sszTyp)
			t.Logf("root: %x\n", root)
		})
	}
}
