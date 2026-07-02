package inspector

import (
        "io"

        "github.com/africanecMorj/mitigation-proxy.git/pkg"
        "golang.org/x/sys/unix"
)

type PostgresMessageType int

const (
	PostgresUnknown PostgresMessageType = iota
	PostgresSSLRequest
	PostgresCancelRequest
	PostgresStartupMessage
)


type Postgres struct {
	buf      []byte

	msgType  PostgresMessageType

	user     string
	database string
	params   map[string]string
}


func (p *Postgres) Read(fd int) (bool, error) {
	tmp := make([]byte, 4096)

	for {
		n, err := unix.Read(fd, tmp[:])

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

		p.buf = append(p.buf, tmp[:n]...)

		if len(p.buf) < 8 {
			continue
		}

		length :=
			int(p.buf[0])<<24 |
				int(p.buf[1])<<16 |
				int(p.buf[2])<<8 |
				int(p.buf[3])


		if length < 8 {
			return false, pkg.ErrNotPostgres
		}

		if len(p.buf) < length {
			continue
		}


		version :=
			int(p.buf[4])<<24 |
				int(p.buf[5])<<16 |
				int(p.buf[6])<<8 |
				int(p.buf[7])

		switch version {

		case 80877103:
			p.msgType = PostgresSSLRequest
			return true, nil

		case 80877102:
			p.msgType = PostgresCancelRequest
			return true, nil

		case 196608:
			p.msgType = PostgresStartupMessage

			info, err := pkg.ParsePostgresStartup(
				p.buf[:length],
			)

			if err != nil {
				return false, err
			}

			p.user = info.User
			p.database = info.Database
			p.params = info.Params

			return true, nil

		default:
			return false, pkg.ErrNotPostgres
		}
	}
}

func NewPostgres() Inspector {

	return &Postgres{
		buf: acquirePreBuf(),

		params: make(map[string]string),
	}
}



func (p *Postgres) RouteKey() RouteInfo {
	meta := make(map[string]string)

	for k, v := range p.params {
		meta[k] = v
	}

	switch p.msgType {

	case PostgresSSLRequest:
		meta["pg_type"] = "ssl_request"

	case PostgresCancelRequest:
		meta["pg_type"] = "cancel_request"

	case PostgresStartupMessage:
		meta["pg_type"] = "startup"
	}

	if p.user != "" {
		meta["user"] = p.user
	}

	if p.database != "" {
		meta["database"] = p.database
	}

	return RouteInfo{
		Protocol: PostgresProto,
		Meta:     meta,
	}
}


func (p *Postgres) Data() []byte {
        return p.buf
}



func (p *Postgres) Close(){

        releasePreBuf(p.buf)

        p.buf=nil
}


