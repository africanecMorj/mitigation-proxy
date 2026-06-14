package transport

import (
	"sync"

	"golang.org/x/sys/unix"
)

const pipeSize = 1 << 20

type Pipe struct {
	r int
	w int
}

//pipe
var pipePool = sync.Pool{
	New: func() any {
		fds := make([]int, 2)
		if err := unix.Pipe(fds); err != nil {
			panic(err)
		}

		unix.FcntlInt(uintptr(fds[0]), unix.F_SETPIPE_SZ, pipeSize)

		return &Pipe{r: fds[0], w: fds[1]}
	},
}

func acquirePipe() *Pipe {
	return pipePool.Get().(*Pipe)
}

func releasePipe(p *Pipe) {
	pipePool.Put(p)
}

//conn
var connPool = sync.Pool{
	New: func() any {
		return new(Conn)
	},
}

func acquireConn() *Conn {
	return connPool.Get().(*Conn)
}

func releaseConn(c *Conn) {
	resetConn(c)
	connPool.Put(c)
}
