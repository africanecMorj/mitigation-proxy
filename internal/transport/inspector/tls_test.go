package inspector

import (
	"testing"
	"io"
	"errors"

	"golang.org/x/sys/unix"
)

func TestTLSEOF(t *testing.T) {
	fds, err := unix.Socketpair(
		unix.AF_UNIX,
		unix.SOCK_STREAM,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer unix.Close(fds[0])

	unix.Close(fds[1])

	tls := NewTLS()

	_, err = tls.Read(fds[0])

	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestTLSRouteKey(t *testing.T) {
	tls := &TLS{
		sni: "example.com",
		alpn: []string{"h2", "http/1.1"},
	}

	info := tls.RouteKey()

	if info.Protocol != TLSProto {
		t.Fatal("invalid protocol")
	}

	if info.Host != "example.com" {
		t.Fatal("invalid host")
	}

	if len(info.ALPN) != 2 {
		t.Fatal("invalid alpn")
	}
}

func TestTLSData(t *testing.T) {
	tls := &TLS{
		buf: []byte{1, 2, 3},
	}

	got := tls.Data()

	if len(got) != 3 {
		t.Fatal("invalid data")
	}
}

func TestTLSClose(t *testing.T) {
	tls := NewTLS()

	tls.buf = append(tls.buf, []byte("hello")...)

	tls.Close()

	if tls.buf != nil {
		t.Fatal("buffer not released")
	}
}

func TestTLSIncompleteRecord(t *testing.T) {
	fds, err := unix.Socketpair(
		unix.AF_UNIX,
		unix.SOCK_STREAM,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	record := []byte{
		0x16,
		0x03,
		0x03,
		0x00,
		0x20,
	}

	_, err = unix.Write(fds[1], record)
	if err != nil {
		t.Fatal(err)
	}

	unix.Close(fds[1])

	tls := NewTLS()

	done, err := tls.Read(fds[0])

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}

	if done {
		t.Fatal("expected incomplete record")
	}
}

func TestTLSAccumulation(t *testing.T) {
	tls := NewTLS()

	tls.buf = append(
		tls.buf,
		0x16,
		0x03,
		0x03,
	)

	if len(tls.buf) != 3 {
		t.Fatal("invalid buffer")
	}

	tls.buf = append(
		tls.buf,
		0x00,
		0x20,
	)

	if len(tls.buf) != 5 {
		t.Fatal("invalid accumulation")
	}
}