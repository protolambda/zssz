package zssz

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
	. "zssz/types"
)

type getSSZFn func() (SSZ, error)

func booltest() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(true))
}
func uint8test() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(uint8(0)))
}
func uint16test() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(uint16(0)))
}
func uint32test() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(uint32(0)))
}
func uint64test() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(uint64(0)))
}

type fixedTestStruct struct {
	a uint8
	b uint64
	c uint32
}
func fixedTestStructTest() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(new(fixedTestStruct)).Elem())
}

type varTestStruct struct {
	a uint16
	b []uint16
	c uint8
}

func varTestStructTest() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(new(varTestStruct)).Elem())
}

type complexTestStruct struct {
	a uint16
	b []uint16
	c uint8
	d []byte
	e varTestStruct
	f [4]fixedTestStruct
	g [2]varTestStruct
}

func complexTestStructTest() (SSZ, error) {
	return SSZFactory(reflect.TypeOf(new(complexTestStruct)).Elem())
}

// note: expected strings are in little-endian, hence the seemingly out of order bytes.
var testCases = []struct {
	// name of test
	name string
	// any value
	value interface{}
	// hex formatted, no prefix
	expected string
	// ssz typ getter
	getSSZ getSSZFn
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
	{"fixedTestStruct", fixedTestStruct{a: 0xab, b: 0xaabbccdd00112233, c: 0x12345678}, "ab33221100ddccbbaa78563412", fixedTestStructTest},
	{"varTestStruct nil", varTestStruct{a: 0xabcd, b: nil, c: 0xff}, "cdab07000000ff", varTestStructTest},
	{"varTestStruct empty", varTestStruct{a: 0xabcd, b: make([]uint16, 0), c: 0xff}, "cdab07000000ff", varTestStructTest},
	{"varTestStruct some", varTestStruct{a: 0xabcd, b: []uint16{1, 2, 3}, c: 0xff}, "cdab07000000ff010002000300", varTestStructTest},
	{"complexTestStruct", complexTestStruct{
		a: 0xaabb,
		b: []uint16{0x1122, 0x3344},
		c: 0xff,
		d: []byte("foobar"),
		e: varTestStruct{a: 0xabcd, b: []uint16{1, 2, 3}, c: 0xff},
		f: [4]fixedTestStruct{
			{0xcc,0x4242424242424242,0x13371337},
			{0xdd,0x3333333333333333,0xabcdabcd},
			{0xee,0x4444444444444444,0x00112233},
			{0xff,0x5555555555555555,0x44556677}},
		g: [2]varTestStruct{
			{a: 0xabcd, b: []uint16{1, 2, 3}, c: 0xff},
			{a: 0xabcd, b: []uint16{1, 2, 3}, c: 0xff}},
	},
	"bbaa" +
		"47000000" + // offset of b, []uint16
		"ff" +
		"4b000000" + // offset of foobar
		"51000000" + // offset of e
		"cc424242424242424237133713" +
		"dd3333333333333333cdabcdab" +
		"ee444444444444444433221100" +
		"ff555555555555555577665544" +
		"5e000000" + // pointer to g
		"22114433" + // contents of b
		"666f6f626172" + // foobar
		"cdab07000000ff010002000300" + // contents of e
		"08000000" + "15000000" + // [start g]: local offsets of [2]varTestStruct
		"cdab07000000ff010002000300" +
		"cdab07000000ff010002000300", complexTestStructTest},
}

func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := tt.getSSZ()
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
			if res := fmt.Sprintf("%x", data); res != tt.expected {
				t.Errorf("got %s, expected %s", res, tt.expected)
			}
		})
	}
}
