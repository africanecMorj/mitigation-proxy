package admin

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/runtime"
)

const SocketPath = "/tmp/mitigation.sock"

type Server struct {
	ln net.Listener
	rt *runtime.Runtime
}

func Start(rt *runtime.Runtime) (*Server, error) {
	if _, err := os.Stat(SocketPath); err == nil {

		conn, err := net.DialTimeout(
			"unix",
			SocketPath,
			time.Second,
		)

		if err == nil {
			conn.Close()
			return nil, fmt.Errorf("mitigation-proxy is already running")
		}

		// stale socket
		if err := os.Remove(SocketPath); err != nil {
			return nil, err
		}
	}

	ln, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(SocketPath, 0600); err != nil {
		ln.Close()
		os.Remove(SocketPath)
		return nil, err
	}

	s := &Server{
		ln: ln,
		rt: rt,
	}

	go s.run()

	return s, nil
}

func (s *Server) run() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			continue
		}

		go handle(conn, s.rt)
	}
}

func (s *Server) Close() error {
	err := s.ln.Close()

	if removeErr := os.Remove(SocketPath); removeErr != nil &&
		!os.IsNotExist(removeErr) {
		return removeErr
	}

	return err
}