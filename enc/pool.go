package enc

import "sync"

// TODO: maybe make buffer-pools ssz-type specific,
//  and initialize buffers with the known fixed-size of a type, and maybe plus some extra if not fixed length?

// get a cleaned buffer from the pool
func GetPooledBuffer() *EncodingBuffer {
	eb := bufferPool.Get().(*EncodingBuffer)
	eb.Reset()
	return eb
}

func ReleasePooledBuffer(eb *EncodingBuffer) {
	bufferPool.Put(eb)
}

// Encoding Buffers are pooled.
var bufferPool = sync.Pool{
	New: func() interface{} { return &EncodingBuffer{} },
}
