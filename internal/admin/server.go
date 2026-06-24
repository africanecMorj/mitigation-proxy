package admin

import (
	"net"
	"os"

	"github.com/africanecMorj/mitigation-proxy.git/internal/runtime"
)

func StartServer(rt *runtime.Runtime) error {
	_ = os.Remove(SocketPath)

	ln, err := net.Listen("unix", SocketPath)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}

			go handle(conn, rt)
		}
	}()

	return nil
}