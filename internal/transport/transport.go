package transport

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/africanecMorj/mitigation-proxy.git/internal/logger"

	"golang.org/x/sys/unix"
)

type BackendDialer func() (int, error)

type BackendPicker interface {
	Pick(r *inspector.RouteInfo, clientIP string) (*health.Backend, int, error)
}

type Wrapper struct {
	Picker    BackendPicker
	Inspector InspectorFactory
}

type InspectorFactory func() inspector.Inspector

func NewWrapper(
    proto string,
    p BackendPicker,
) Wrapper {

    switch proto {
    case "tls":
        return Wrapper{
            Inspector: inspector.NewTLS,
            Picker:    p,
        }

    case "http":
        return Wrapper{
            Inspector: inspector.NewHTTP,
            Picker:    p,
        }

    case "postgres":
        return Wrapper{
            Inspector: inspector.NewPostgres,
            Picker:    p,
        }

    case "quic":
        return Wrapper{
            Inspector: inspector.NewQUIC,
            Picker:    p,
        }

    default:
        return Wrapper{
            Inspector: inspector.NewTCP,
            Picker:    p,
        }
    }
}

type Transport struct {
    id         uint64
    listenerFD int
    loop       *EventLoop
    logger     logger.Logger
}

func New(
    w *Wrapper,
    l logger.Logger,
) (*Transport, error) {

    loop, err := NewEventLoop(w, l)
    if err != nil {
        return nil, err
    }

    l.Info(
        "transport created",
        map[string]interface{}{
            "transport_id": logger.NextID(),
        },
    )

    return &Transport{
        id:     logger.NextID(),
        loop:   loop,
        logger: l,
    }, nil
}

func (t *Transport) Run(listenerFD int) error {
    t.listenerFD = listenerFD

    t.logger.Info(
        "transport started",
        map[string]interface{}{
            "transport_id": t.id,
            "listener_fd": listenerFD,
        },
    )

    err := t.loop.Run(listenerFD)

    if err != nil {
        t.logger.Error(
            "transport stopped",
            map[string]interface{}{
                "transport_id": t.id,
                "error": err.Error(),
            },
        )
    }

    return err
}

func (t *Transport) Reload(w *Wrapper) {
    t.logger.Info(
        "transport reload",
        map[string]interface{}{
            "transport_id": t.id,
        },
    )

    t.loop.picker.Store(w.Picker)
    t.loop.inspector.Store(w.Inspector)
}

func (t *Transport) Close() {
    t.logger.Info(
        "transport closing",
        map[string]interface{}{
            "transport_id": t.id,
            "fd": t.listenerFD,
        },
    )

    unix.Close(t.listenerFD)
}