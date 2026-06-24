package inspector

import (
	"github.com/africanecMorj/mitigation-proxy.git/pkg"

	"io"

	"golang.org/x/sys/unix"
)



type TLS struct {
	buf  []byte
	sni  string
	alpn []string
}


func (t *TLS) Read(fd int) (bool, error) {
	tmp := make([]byte, 4096)

	for {
		n, err := unix.Read(fd, tmp)

		if err != nil {

			if err == unix.EINTR {
				continue
			}

			if err == unix.EAGAIN {
				return false, nil
			}
			
			return false, err
		}

		if n == 0 {
			return false, io.EOF
		}

		t.buf = append(t.buf, tmp[:n]...)

		if len(t.buf) >= 5 {
			length := int(t.buf[3])<<8 | int(t.buf[4])
			total := 5 + length

			if len(t.buf) >= total {
				hello, err := pkg.ParseClientHello(t.buf[:total])
				if err == nil {
					t.sni = hello.SNI
					t.alpn = hello.ALPN
					
				}
			
				return true, nil
				
			}
		}
	}
}

func NewTLS() *TLS {
	return &TLS{
		buf: acquirePreBuf(),
	}
}

func (t *TLS) RouteKey() RouteInfo {
	return RouteInfo {
		SNI: t.sni,
		ALPN: t.alpn,
	}
}

func (t *TLS) Close() {
    releasePreBuf(t.buf)
    t.buf = nil
}

func (t *TLS) Data() []byte {
	return t.buf
}

