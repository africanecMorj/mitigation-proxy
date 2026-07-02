package transport

import (
	"testing"
	"sync/atomic"

	"github.com/africanecMorj/mitigation-proxy.git/pkg"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/africanecMorj/mitigation-proxy.git/internal/logger"

	"golang.org/x/sys/unix"
)

type mockPicker struct {
	backend *health.Backend
	fd      int
	err     error
}

func (m mockPicker) Pick(
	_ *inspector.RouteInfo,
	_ string,
) (*health.Backend, int, error) {
	return m.backend, m.fd, m.err
}

type mockInspector struct {
	done bool
	err  error

	info inspector.RouteInfo
	data []byte
}

func (m *mockInspector) Read(fd int) (bool, error) {
	return m.done, m.err
}

func (m *mockInspector) RouteKey() inspector.RouteInfo {
	return m.info
}

func (m *mockInspector) Data() []byte {
	return m.data
}

func (m *mockInspector) Close() {}

func BenchmarkAcceptLoop(b *testing.B) {
	fd, err := pkg.BuildListener("127.0.0.1:0")
	_, sa, err := pkg.ResolveSockaddr("127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}

	defer unix.Close(fd)

	l := &EventLoop{
		conns: make(map[int]*Conn),
	}

	go func() {
		for {
			cfd, _ := unix.Socket(
				unix.AF_INET,
				unix.SOCK_STREAM,
				0,
			)

			unix.Connect(cfd, sa)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.acceptLoop(fd)
	}
}

func BenchmarkAtomicPickerLoad(b *testing.B) {
	var v atomic.Value

	v.Store(mockPicker{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.Load().(BackendPicker)
	}
}

