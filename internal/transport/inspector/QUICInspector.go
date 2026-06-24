package inspector

import (
	"io"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/pkg"
)

type QUIC struct {
	buf  []byte
	sni  string
	alpn []string
}

func (q *QUIC) Read(fd int) (bool, error) {

	tmp := make([]byte, 2048)

	for {

		n, err := unix.Read(fd, tmp)

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


		q.buf = append(q.buf, tmp[:n]...)


		// QUIC long header minimum
		if len(q.buf) < 6 {
			continue
		}


		if !isQUICInitial(q.buf) {
			continue
		}


		clientHello, err := pkg.ParseQUICClientHello(q.buf)

		if err != nil {
			return false, err
		}


		q.sni = clientHello.SNI
		q.alpn = clientHello.ALPN


		return true, nil
	}
}


func isQUICInitial(pkt []byte) bool {

	// QUIC long header bit
	if pkt[0]&0x80 == 0 {
		return false
	}


	// Packet type bits
	packetType := (pkt[0] >> 4) & 0x03


	// Initial packet type = 0
	return packetType == 0
}



func NewQUIC() *QUIC {

	return &QUIC{
		buf: acquirePreBuf(),
	}
}



func (q *QUIC) RouteKey() RouteInfo {

	return RouteInfo{
		SNI:  q.sni,
		ALPN: q.alpn,
	}
}



func (q *QUIC) Close() {

	releasePreBuf(q.buf)

	q.buf = nil
}



func (q *QUIC) Data() []byte {

	return q.buf
}