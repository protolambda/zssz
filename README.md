# ZSSZ

Highly optimized SSZ encoder. Uses unsafe Go, safely.

"ZSSZ", a.k.a. ZRNT-SSZ, is the SSZ version used and maintained for [ZRNT](https://github.com/protolambda/zrnt),
 the ETH 2.0 Go executable spec.


Features:
- Zero-allocations where possible
   - offset checking allocates a small array of `uint32`s
   - dynamic size uses buffer pooling to encode dynamic data *in order*
   - small value passing is preferred over passing slices, avoid memory on the heap. 
- Construct all encoding/decoding/hashing logic for a type, then run it 10000 times the efficient way
- No reflection during encoding/decoding/hashing execution of the constructed SSZ-type
    - Exception: slice allocation uses `reflect.MakeSlice`, and pointer-value allocation uses `reflect.New`,
      but the type is already readily available.
      This is to avoid the GC collecting allocated space within a slice, and be safe on memory assumptions.
- Construction of SSZ types can also be used to support encoding of dynamic types
- No dependencies other than the standard Go libs.
    - Zero-hashes are pre-computed with the `sha256` package,
       but you can supply a more efficient version to run hash-tree-root with. 
       (E.g. reduce allocations by re-using a single state)
- Hardened: A work in progress now, but all SSZ rules are strictly yet efficiently enforced.
- Fuzzmode-decoding: decode arbitrary data into a struct.
  The length of the input + contents determine the length of dynamic parts.
- Replaceable hash-function. Initialize the pre-computed zero-hashes with `InitZeroHashes(yourHashFn)`
  and then call `HashTreeRoot(yourHashFn, val, sszType)`. Zero-hashes default to SHA-256.
- Passes the Eth 2.0 Static-SSZ tests, in the [ZRNT](https://github.com/protolambda/zrnt) test suite.

Supported types
- small basic-types (`bool`, `uint8`, `uint16`, `uint32`, `uint64`)
- containers
  - incl. optimizations for fixed-size only elements
- vector
  - `Vector`, optimized for fixed and variable size elements
  - `BytesN`, optimized for bytes (fixed length)
  - `BasicVector`, optimized for basic element types
- list
  - `List` optimized for fixed and variable size elements
  - `Bytes`, optimized for bytes (dynamic length)
  - `BasicList`, optimized for basic element types

Possibly supported in future:
- union/null
- uint128/uint256
- strings
- partials

TODO:
- benchmarking. How does it compare to SSZ using reflection? And to the golang-serialization, Gob?
- support for SSZ caching


## Usage

Here is an example that illustrates the full usage:
```go
package main

import (
	. "github.com/protolambda/zssz/types"
	. "github.com/protolambda/zssz/htr"
	. "github.com/protolambda/zssz"
	"reflect"
	"bytes"
	"bufio"
	"fmt"
	"crypto/sha256"
)

type Bar struct {
	DogeImg []byte
	Timestamp uint64
	ID [32]byte
}

type MyThing struct {
	Foo uint32
	Bars [2]Bar
	SomeSignature [96]byte
}

func main() {
	myThingSSZ := GetSSZ((*MyThing)(nil))

	// example instance filled with dummy data
	obj := MyThing{
		Foo: 123,
		Bars: [2]Bar{
			{DogeImg: []byte{1, 2, 3}, Timestamp: 0x112233, ID: [32]byte{1}},
			{DogeImg: []byte{7, 8, 9}, Timestamp: 12345678, ID: [32]byte{2}},
		},
		SomeSignature: [96]byte{0,1,2,3,4,5,6,7,8,9},
	}

	// encoding
	// -----------------------
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	// Note that Encode takes any io.Writer
	if err := Encode(bufWriter, &obj, myThingSSZ); err != nil {
		panic(err)
	}
	if err := bufWriter.Flush(); err != nil {
		panic(err)
	}
	data := buf.Bytes()
	fmt.Printf("encoded myThing: %x\n", data)

	// decoding
	// -----------------------
	dst := MyThing{}
	bytesLen := uint32(len(data))
    r := bytes.NewReader(data)
	// note that Decode takes any io.Reader
	if err := Decode(r, bytesLen, &dst, myThingSSZ); err != nil {
		panic(err)
	}
	fmt.Printf("decoded myThing: some data from it: %v\n", dst.Bars[1].DogeImg[:])

	// hash-tree-root
	// -----------------------
	// get the root
	root := HashTreeRoot(sha256.Sum256, &obj, myThingSSZ)
	fmt.Printf("root of my thing: %x\n", root)
	
	signingRoot := SigningRoot(sha256.Sum256, &obj, myThingSSZ.(SignedSSZ))
	fmt.Printf("signing-root of my thing: %x\n", signingRoot)
}
```

## Extending

ZRNT-SSZ does not check for interfaces on types to do custom SSZ encoding etc., as this would be to slow and inflexible.

Instead, it gives you the freedom to compose your own custom SSZ type-factory,
 the function that is used to compose a `SSZ` structure.

The default factory function is building the structure when calling `GetSSZ((*SomeType)(nil))`

With this, you can remove/change existing functionality, and add your own. 
Pure composition, without performance degradation 
(composition happens before actual execution of SSZ serialization/hashing).


### Composition example

```go
package main

import (
	. "github.com/protolambda/zssz/types"
	"reflect"
)

// A simple interface to call when composing a SSZ for your type.
func MyFactory(typ reflect.Type) (SSZ, error) {
	// Pass this same factory to it, this
	return myFactory(MyFactory, typ)
}

// other factories may want to use this one for some of their functionality, make it public.
func MyFactoryFn(factory SSZFactoryFn, typ reflect.Type) (SSZ, error) {
	// figure out which SSZ type definition to use, based on interface check
	if typ.Implements(myFancyCustomTypeInterface) {
		// pass it the typ, it's recommended to make the function check if the type is really allowed.
		// This makes usage by other 3rd-party factories safe.
		return FancyTypeSSZ(typ)
	}
	
	// figure out which SSZ type definition to use, based on kind
	switch typ.Kind() {
	case reflect.Struct:
		// pass it a factory, to build SSZ definitions for child-types (container fields).
		return MyContainerSSZDefinition(factory, typ)
	default:
		// use the default behavior otherwise
		return DefaultSSZFactory(factory, typ)
	}
}

func main() {
	// building a SSZ type definition using a factory
	myThingSSZ, _ := MySSZFactory(reflect.TypeOf((*MyThing)(nil)).Elem())
}
```

### Writing a custom SSZ type definition

Simply implement the `SSZ` interface. And provide some function to instantiate it (See `Vector` as example),
 or maybe declare a singleton instance (See `ssz_basic` as example).
You can then mix it in however you like with the factory pattern, and start serializing/hashing your way.


## Contact

Dev: [@protolambda on Twitter](https://twitter.com/protolambda)


## License

MIT, see license file.

