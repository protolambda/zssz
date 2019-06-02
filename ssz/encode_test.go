package ssz

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
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
	{"uint8 00", uint8(0x00), "00", uint8test},
	{"uint8 ab", uint8(0xab), "ab", uint8test},
	{"uint16 0000", uint16(0x0000), "0000", uint16test},
	{"uint16 abcd", uint16(0xabcd), "cdab", uint16test},
	{"uint32 00000000", uint32(0x00000000), "00000000", uint32test},
	{"uint32 01234567", uint32(0x01234567), "67452301", uint32test},
	{"uint64 0000000000000000", uint64(0x00000000), "0000000000000000", uint64test},
	{"uint64 0123456789abcdef", uint64(0x0123456789abcdef), "efcdab8967452301", uint64test},
}

func TestSSZEncode(t *testing.T) {
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
