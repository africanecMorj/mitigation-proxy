package transport

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"

	"golang.org/x/sys/unix"
)

type BackendDialer func() (int, error)

type BackendPicker interface {
	Pick(sni, host string, alpn []string, clientIP string) (*health.Backend, int, error)
}

type Wrapper struct {
	Picker    BackendPicker
	Inspector inspector.Inspector
}

func NewWrapper(
    proto string,
    p BackendPicker,
) Wrapper {

    switch proto {
    case "tls":
        return Wrapper{
            Inspector: inspector.NewTLS(),
            Picker:    p,
        }

    case "http":
        return Wrapper{
            Inspector: inspector.NewHTTP(),
            Picker:    p,
        }

    case "quic":
        return Wrapper{
            Inspector: inspector.NewQUIC(),
            Picker:    p,
        }

    default:
        return Wrapper{
            Inspector: inspector.NewTCP(),
            Picker:    p,
        }
    }
}

type Transport struct {
    listenerFD int
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
	t.listenerFD = listenerFD
    return t.loop.Run(listenerFD)
}

func (t *Transport) Reload(
	w *Wrapper,
) {
	t.loop.picker.Store(w.Picker)
	t.loop.inspector.Store(w.Inspector)
}

func (t *Transport) Close() {
	unix.Close(t.listenerFD)
}