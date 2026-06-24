package inspector

type Protocol int32 

const (
    TLSProto = Protocol iota
    PostgresProto
    HTTPProto
    RawTCPProto
)



type RouteInfo struct {
    SNI  string
    ALPN []string
    Host string
  
    Metadata map[string]string
}

type Inspector interface {
    Read(fd int) (bool, error)
    RouteKey() RouteInfo
    Data() []byte
    Close()
}
