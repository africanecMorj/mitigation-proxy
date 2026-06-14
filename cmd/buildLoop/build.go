package buildLoop

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
)

func Build(cfg *config.Config) error {
	clusters, err := config.BuildClusters(cfg)
	if err != nil {
		return err
	}

	for _, l := range cfg.Listeners {
		fd, err := unix.Socket(
			unix.AF_INET,
			unix.SOCK_STREAM|unix.SOCK_NONBLOCK,
			0,
		)
		if err != nil {
			return err
		}

		unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)

		host, portStr, err := net.SplitHostPort(l.Address)
		if err != nil {
			return err
		}

		port, _ := strconv.Atoi(portStr)

		ip := net.ParseIP(host).To4()

		addr := &unix.SockaddrInet4{
			Port: port,
		}
		copy(addr.Addr[:], ip)

		if err := unix.Bind(fd, addr); err != nil {
			return err
		}

		if err := unix.Listen(fd, 1024); err != nil {
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

		go tr.Run(fd)

	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh 

	return nil

}
