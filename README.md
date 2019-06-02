# ZRNT-SSZ

Highly optimized SSZ encoder. Uses unsafe Go, safely.

Features:
- Zero-allocations where possible
   - offset checking allocates a small array of `uint32`s
   - dynamic size uses buffer pooling to encode dynamic data *in order*
   - small value passing is preferred over passing slices, avoid memory on the heap. 
- Construct all encoding/decoding/hashing logic for a type, then run it 10000 times the efficient way
- No reflection during encoding/decoding/hashing execution of the constructed SSZ-type
- Construction of SSZ types can also be used to support encoding of dynamic types


**work in progress**

TODO common:
- [x] Unsafe util to work with arrays/slices
- [x] Encoding buffer, with functions tweaked for SSZ-readability
- [x] Buffer pool
- [x] SSZ type interface
- [x] typ constructor
- [x] encoder interface
- [x] decoder interface
- [x] hash-tree-root interface
- [x] signing-root interface

TODO type-pre-compiling:
- [x] small basic-types
- [ ] uint128/uint256
- [x] containers
- [x] bytesN
- [x] vector
- [x] bytes
- [x] list

TODO encoding:
- [x] small basic-types
- [ ] uint128/uint256
- [x] containers
- [x] bytesN
- [x] vector
- [x] bytes
- [x] list
- [ ] union/null

TODO decoding:
- [x] small basic-types
- [ ] uint128/uint256
- [x] containers
- [x] bytesN
- [x] vector
- [x] bytes
- [x] list
- [ ] union/null

TODO hash-tree-root:
- [x] small basic-types
- [ ] uint128/uint256
- [x] containers
- [x] bytesN
- [ ] vector
- [x] bytes
- [x] list
- [ ] union/null

TODO testing:
- [ ] pass spec tests
- [ ] benchmarking. How does it compare to SSZ using reflection? And to the golang-serialization, Gob?

## Usage

Here is an example that illustrates the full usage:
```go
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
	// building a SSZ type definition
	myThingSSZ, _ := SSZFactory(reflect.TypeOf(new(MyThing)))

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
	fmt.Printf("encoded myThing: %x\n", buf.Bytes())

	// decoding
	// -----------------------
	dst := MyThing{}
	// note that Decode takes any io.Reader
	if err := Decode(&buf, &dst, myThingSSZ); err != nil {
		panic(err)
	}
	fmt.Printf("decoded myThing: some data from it: %v\n", dst.Bars[1].DogeImg[:])

	// hash-tree-root
	// -----------------------
	// simple wrapper around SHA-256. You can make it re-use its state. (hash.Reset())
	hashFn := func(input []byte) (out []byte) {
		hash := sha256.New()
		hash.Write(input)
		return hash.Sum(nil)
	}
	// get the root
	root := HashTreeRoot(hashFn, &obj, myThingSSZ)
	fmt.Printf("root of my thing: %x\n", root)
	
	signingRoot := SigningRoot(hashFn, &obj, myThingSSZ.(SignedSSZ))
	fmt.Printf("signing-root of my thing: %x\n", signingRoot)
}
```

## Extending

ZRNT-SSZ does not check for interfaces on types to do custom SSZ encoding etc., as this would be to slow and inflexible.

Instead, it gives you the freedom to compose your own custom SSZ type-factory,
 the function that is used to compose a `SSZ` structure.

With this, you can remove/change existing functionality, and add your own. 
Pure composition, without performance degradation 
(composition happens before actual execution of SSZ serialization/hashing).


### Composition example

```go
package main

import (
	. "github.com/protolambda/zrnt-ssz/ssz"
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
```

### Writing a custom SSZ type definition

Simply implement the `SSZ` interface. And provide some function to instantiate it (See `Vector` as example),
 or maybe declare a singleton instance (See `ssz_basic` as example).
You can then mix it in however you like with the factory pattern, and start serializing/hashing your way.


## Contact

Dev: [@protolambda on Twitter](https://twitter.com/protolambda)


## License

MIT, see license file.

