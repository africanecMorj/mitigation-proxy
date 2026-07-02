package inspector

type TCP struct{}

func NewTCP() Inspector {
	return &TCP{}
}

func (t *TCP) Read(fd int) (bool, error) {
	return true, nil
}

func (t *TCP) RouteKey() RouteInfo {
	return RouteInfo{
		Protocol:RawTCPProto,
	}
}

func (t *TCP) Data() []byte {
	return nil
}

func (t *TCP) Close() {}