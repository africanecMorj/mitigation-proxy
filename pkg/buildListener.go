package pkg

import (
	"strconv"
	"net"

	"golang.org/x/sys/unix"
)

func BuildListener (a string) (int, error){
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

