package transport

import (
	"errors"
	"io"

	"golang.org/x/sys/unix"
)

type SpliceResult struct {
	Bytes int64
	
	NeedRead  bool
	NeedWrite bool
	Err       error
}

type Splicer struct {
	pipe *Pipe

	buf int
}


func NewSplicer(p *Pipe) *Splicer {
	return &Splicer{pipe: p}
}

func (s *Splicer) Transfer(src, dst int) SpliceResult {
	var res SpliceResult

	for {
		progress := false
		srcEOF := false

		// fill pipe
		for s.buf < pipeSize {
			n, err := unix.Splice(
				src, nil,
				s.pipe.w, nil,
				pipeSize-s.buf,
				unix.SPLICE_F_MOVE|unix.SPLICE_F_NONBLOCK,
			)

			if errors.Is(err, unix.EINTR) {
				continue
			}

			if errors.Is(err, unix.EAGAIN) {
				if s.buf == 0 {
					res.NeedRead = true
				}
				break
			}

			if err != nil {
				res.Err = err
				return res
			}

			if n == 0 {
				srcEOF = true
				break
			}

			s.buf += int(n)
			res.Bytes += int64(n)
			progress = true
		}

		// flush pipe
		for s.buf > 0 {
			n, err := unix.Splice(
				s.pipe.r, nil,
				dst, nil,
				s.buf,
				unix.SPLICE_F_MOVE|unix.SPLICE_F_NONBLOCK,
			)

			if errors.Is(err, unix.EINTR) {
				continue
			}

			if errors.Is(err, unix.EAGAIN) {
				res.NeedWrite = true
				return res
			}

			if err != nil {
				res.Err = err
				return res
			}

			s.buf -= int(n)
			progress = true
		}

		if srcEOF {
			res.Err = io.EOF
			return res
		}

		if !progress {
			return res
		}
	}

}