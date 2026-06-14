package inspector

import "sync"

var preBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 4096)
		return &b
	},
}

func acquirePreBuf() []byte {
	return (*preBufPool.Get().(*[]byte))[:0]
}

func releasePreBuf(b []byte) {
	if cap(b) != 4096 {
		return
	}

	b = b[:0]
	preBufPool.Put(&b)
}