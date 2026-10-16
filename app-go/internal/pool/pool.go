package pool

import (
	"bytes"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 1024*1024 { // Prevent retaining oversized buffers (>1MB)
		return
	}
	buf.Reset()
	bufferPool.Put(buf)
}

var byteSlicePool256 = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 256)
		return &b
	},
}

func GetByteSlice256() *[]byte {
	return byteSlicePool256.Get().(*[]byte)
}

func PutByteSlice256(b *[]byte) {
	byteSlicePool256.Put(b)
}
