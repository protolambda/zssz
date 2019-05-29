# ZRNT-SSZ

Highly optimized, yet readable, SSZ encoder. Uses unsafe Go, safely.

Features:
- Zero-allocations where possible
   - offset checking allocates a small array of `uint32`s
   - dynamic size uses buffer pooling to encode dynamic data *in order*
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
- [ ] hash-tree-root interface

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

TODO decoding:
- [x] small basic-types
- [ ] uint128/uint256
- [x] containers
- [x] bytesN
- [x] vector
- [x] bytes
- [x] list

TODO hashing:
- [ ] small basic-types
- [ ] uint128/uint256
- [ ] containers
- [ ] bytesN
- [ ] vector
- [ ] bytes
- [ ] list

TODO testing:
- [ ] pass spec tests
- [ ] benchmarking. How does it compare to SSZ using reflection? And to the golang-serialization, Gob?

## Contact

Core dev: [@protolambda on Twitter](https://twitter.com/protolambda)

## License

MIT, see license file.

