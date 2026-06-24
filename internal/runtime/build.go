package runtime

import (
	"net"
	"strconv"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
)

func (rt *Runtime) Build(cfg *config.Config) error {
	clusters, err := config.BuildClusters(cfg)
	if err != nil {
		return err
	}

	rt.RegisterClusters(clusters)

	for _, balancer := range clusters {
		for _, backend := range balancer.Backends() {
			rt.wg.Add(1)
			go rt.WatchBackend(balancer, backend)
		}
	}

	for _, l := range cfg.Listeners {
		
		fd, err := buildListener(l.Address)
		if err != nil {
			return err
		}

		p, err := config.NewPicker(l, clusters)
		if err != nil {
			return err
		}

		w := transport.NewWrapper(
    		l.Routing.Type,
    		p,
		)

		tr, err := transport.New(&w)
		if err != nil {
			return err
		}

		
		rt.Register(l.Address, tr)
		go tr.Run(fd)

	}


	return nil

}

func buildListener (a string) (int, error){
		fd, err := unix.Socket(
			unix.AF_INET,
			unix.SOCK_STREAM|unix.SOCK_NONBLOCK,
			0,
		)
		if err != nil {
			return 0, err
		}

		unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)

		host, portStr, err := net.SplitHostPort(a)
		if err != nil {
			return 0, err
		}

		port, _ := strconv.Atoi(portStr)

		ip := net.ParseIP(host).To4()

		addr := &unix.SockaddrInet4{
			Port: port,
		}
		copy(addr.Addr[:], ip)

		if err := unix.Bind(fd, addr); err != nil {
			return 0, err
		}

		if err := unix.Listen(fd, 1024); err != nil {
			return 0, err
		}

		return fd, nil
}