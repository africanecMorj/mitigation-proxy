package inspector

import (
	"golang.org/x/sys/unix"

	"log"
	"bytes"
	"io"
	"errors"
)

const (
	maxHeaderSize = 64 * 1024
	headerEnd     = "\r\n\r\n"
)

type HTTP struct {
	host string
	buf []byte
}

func (h *HTTP) Read(fd int) (bool, error) {
	var tmp [4096]byte

	for {
		n, err := unix.Read(fd, tmp[:])

		if err != nil {
			switch err {
			case unix.EINTR:
				continue

			case unix.EAGAIN:
				return false, nil

			default:
				return false, err
			}
		}

		if n == 0 {
			return false, io.EOF
		}

		h.buf = append(h.buf, tmp[:n]...)

		if len(h.buf) > maxHeaderSize {
			return false, errors.New("buffer size exceeded")
		}

		idx := bytes.Index(h.buf, []byte(headerEnd))

		if idx == -1 {
			continue
		}

		host, err := parseHost(h.buf[:idx])
		if err != nil {
			return false, err
		}

		h.host = host
		return true, nil
	}
}

func parseHost(headers []byte) (string, error) {
	var host string
	found := false

	for len(headers) > 0 {
		lineEnd := bytes.IndexByte(headers, '\n')

		var line []byte

		if lineEnd == -1 {
			line = headers
			headers = nil
		} else {
			line = headers[:lineEnd]
			headers = headers[lineEnd+1:]
		}

		line = bytes.TrimSuffix(line, []byte("\r"))

		if len(line) < 5 {
			continue
		}

		if !bytes.EqualFold(line[:5], []byte("host:")) {
			continue
		}

		if found {
			return "", errors.New("duplicate host header")
		}

		host = string(bytes.TrimSpace(line[5:]))
		found = true
	}

	if !found {
		return "", errors.New("missing host header")
	}

	return host, nil
}

func NewHTTP() *HTTP {
	return &HTTP{
		buf: acquirePreBuf(),
	}
}

func (h *HTTP) RouteKey() RouteInfo {
	return RouteInfo {
		Host: h.host,
	}
}

func (h *HTTP) Close() {
    releasePreBuf(h.buf)
    h.buf = nil
}

func (h *HTTP) Data() []byte {
	return h.buf
}

