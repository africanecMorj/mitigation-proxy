package strategies

import (
	"sync/atomic"
	"fmt"
	"slices"

	"github.com/zeebo/xxh3"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
)

type Sticky struct {
	balancers.BaseBalancer
	nodes atomic.Pointer[NodesSnapshot]
}

type Node struct {
	Hash    uint64
	Backend *health.Backend
}

type NodesSnapshot struct {
	Nodes []Node
}

func NewSticky(
	backends []*health.Backend,
) *Sticky {

	const virtualNodes = 100

	nodes := make([]Node, 0, len(backends)*virtualNodes)

	for _, b := range backends {
		for i := 0; i < virtualNodes; i++ {
			key := fmt.Sprintf("%s#%d", b.Address, i)

			nodes = append(nodes, Node{
				Hash:    xxh3.HashString(key),
				Backend: b,
			})
		}
	}

	s := &Sticky{
		BaseBalancer: balancers.NewBaseBalancer(
			backends,
		),
	}

	s.nodes.Store(&NodesSnapshot{nodes})

	return s
}

func (lb *Sticky) Next(ip string) *health.Backend {
    backends := lb.Backends()

    switch len(backends) {
    case 0:
        return nil
    case 1:
        return backends[0]
    }

    snap := lb.nodes.Load()
    clientHash := xxh3.HashString(ip)

    idx, _ := slices.BinarySearchFunc(
        snap.Nodes,
        clientHash,
        func(node Node, hash uint64) int {
            switch {
            case node.Hash < hash:
                return -1
            case node.Hash > hash:
                return 1
            default:
                return 0
            }
        },
    )

    n := len(snap.Nodes)

    for i := 0; i < n; i++ {
        node := snap.Nodes[(idx+i)%n]

        switch node.Backend.StateValue() {
        case health.Draining,
            health.Removed,
            health.Shutdown,
            health.Unhealthy:
            continue
        }

        return node.Backend
    }

    return nil
}
