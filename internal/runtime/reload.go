package runtime

import (
	"time"
	
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

func (rt *Runtime) Reload(cfg *config.Config) error {

    newClusters, err := config.BuildClusters(cfg)
    if err != nil {
        return err
    }

    // drain old
    for _, bl := range rt.clusters {
        for _, backend := range bl.Backends() {
            backend.StartDrain(5 * time.Minute, health.Removed)
        }
    }

    // start watchers for new
    for _, bl := range newClusters {
        for _, backend := range bl.Backends() {
            rt.wg.Add(1)
            go rt.WatchBackend(bl, backend)
        }
    }

	rt.RegisterClusters(newClusters)

    // swap transports
    for _, listener := range cfg.Listeners {

        p, err := config.NewPicker(
            listener,
            newClusters,
        )
        if err != nil {
            return err
        }

        w := transport.NewWrapper(
            listener.Routing.Type,
            p,
        )

        if err := rt.reload(
            listener.Address,
            w.Picker,
            w.Inspector,
        ); err != nil {
            return err
        }
    }

    return nil
}