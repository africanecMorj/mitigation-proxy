package inspector

type Inspector interface {
    Read(fd int) (bool, error)
    RouteKey() RouteInfo
    Data() []byte
    Close()
}

type Protocol int32 

const (
    TLSProto Protocol = iota
    PostgresProto
    HTTPProto
    RawTCPProto
)


type RouteInfo struct {
    Protocol Protocol
    
    ALPN []string
    Host string
  
    Meta map[string]string
}