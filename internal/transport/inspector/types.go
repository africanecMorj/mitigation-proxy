package inspector

type RouteInfo struct {
    SNI  string
    ALPN []string
    Host string
}

type Inspector interface {
    Read(fd int) (bool, error)
    RouteKey() RouteInfo
    Data() []byte
    Close()
}
