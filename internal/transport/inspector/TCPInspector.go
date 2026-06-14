package inspector

type TCP struct{}

func NewTCP() *TCP {
	return &TCP{}
}

func (t *TCP) Read(fd int) (bool, error) {
	return true, nil
}

func (t *TCP) RouteKey() RouteInfo {
	return RouteInfo{}
}

func (t *TCP) Data() []byte {
	return nil
}

func (t *TCP) Close() {}