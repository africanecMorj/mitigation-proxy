package transport

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)

type BackendDialer func() (int, error)

type BackendPicker interface {
	Pick(host string, alpn []string, clientIP string) (*health.Backend, int, error)
}

type Wrapper struct {
	picker    BackendPicker
	inspector inspector.Inspector
}

func NewWrapper(
    proto string,
    p BackendPicker,
) Wrapper {

    switch proto {
    case "tls":
        return Wrapper{
            inspector: inspector.NewTLS(),
            picker:    p,
        }

    case "http":
        return Wrapper{
            inspector: inspector.NewHTTP(),
            picker:    p,
        }

    default:
        return Wrapper{
            inspector: inspector.NewTCP(),
            picker:    p,
        }
    }
}

type Transport struct {
	loop *EventLoop
}

func New(w *Wrapper) (*Transport, error) {
	loop, err := NewEventLoop(w)
	if err != nil {
		return nil, err
	}

	return &Transport{loop: loop}, nil
}

func (t *Transport) Run(listenerFD int) error {
	return t.loop.Run(listenerFD)
}
