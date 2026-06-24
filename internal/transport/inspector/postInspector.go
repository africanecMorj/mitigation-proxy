package inspector

import (
        "io"
        
        "github.com/africanecMorj/mitigation-proxy.git/pkg"
        "golang.org/x/sys/unix"
)

type Postgres struct {
        buf      []byte
        user     string
        database string
        params   map[string]string
}


func (p *Postgres) Read(fd int) (bool, error) {

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


                p.buf = append(p.buf, tmp[:n]...)


                // postgres packet header
                if len(p.buf) >= 4 {

                        length :=
                                int(p.buf[0])<<24 |
                                int(p.buf[1])<<16 |
                                int(p.buf[2])<<8 |
                                int(p.buf[3])


                        if len(p.buf) >= length {


                                startup := p.buf[:length]


                                info, err := pkg.ParsePostgresStartup(startup)

                                if err == nil {

                                        p.user = info.User
                                        p.database = info.Database
                                        p.params = info.Params

                                }


                                return true,nil
                        }
                }
        }
}


func NewPostgres() *Postgres {

        return &Postgres{
                buf: acquirePreBuf(),
                params: make(map[string]string),
        }
}



func (p *Postgres) RouteKey() RouteInfo {

        return RouteInfo{
				Protocol: PostgresProto

				Meta: map[string]string{
					"user": p.user,
                	"database": p.database,
                	"params": p.params
				}
                
        }
}



func (p *Postgres) Data() []byte {
        return p.buf
}



func (p *Postgres) Close(){

        releasePreBuf(p.buf)

        p.buf=nil
}


