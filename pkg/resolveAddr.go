package pkg

import (
	"errors"
	"net"

	"golang.org/x/sys/unix"
)

func ResolveSockaddr(address string) (int, unix.Sockaddr, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return 0, nil, err
	}

	if ip4 := addr.IP.To4(); ip4 != nil {
		sa := &unix.SockaddrInet4{
			Port: addr.Port,
		}
		copy(sa.Addr[:], ip4)

		return unix.AF_INET, sa, nil
	}

	if ip6 := addr.IP.To16(); ip6 != nil {
		sa := &unix.SockaddrInet6{
			Port: addr.Port,
		}
		copy(sa.Addr[:], ip6)

		return unix.AF_INET6, sa, nil
	}

	return 0, nil, errors.New("unsupported ip address")
}