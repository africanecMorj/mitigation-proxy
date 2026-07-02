package inspector

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestHTTPRead(t *testing.T) {
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

	req := []byte(
		"GET / HTTP/1.1\r\n" +
			"Host: example.com\r\n" +
			"\r\n",
	)

	go unix.Write(fds[1], req)

	h := NewHTTP()

	done, err := h.Read(fds[0])

	if err != nil {
		t.Fatal(err)
	}

	if !done {
		t.Fatal("expected done=true")
	}

	info := h.RouteKey()

	if info.Host != "example.com" {
		t.Fatalf(
			"got %q want example.com",
			info.Host,
		)
	}
}

func TestHTTPMissingHost(t *testing.T) {
	fds, _ := unix.Socketpair(
		unix.AF_UNIX,
		unix.SOCK_STREAM,
		0,
	)

	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	req := []byte(
		"GET / HTTP/1.1\r\n" +
			"User-Agent: test\r\n" +
			"\r\n",
	)

	go unix.Write(fds[1], req)

	h := NewHTTP()

	_, err := h.Read(fds[0])

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHTTPDuplicateHost(t *testing.T) {
	fds, _ := unix.Socketpair(
		unix.AF_UNIX,
		unix.SOCK_STREAM,
		0,
	)

	defer unix.Close(fds[0])
	defer unix.Close(fds[1])

	req := []byte(
		"GET / HTTP/1.1\r\n" +
			"Host: a.com\r\n" +
			"Host: b.com\r\n" +
			"\r\n",
	)

	go unix.Write(fds[1], req)

	h := NewHTTP()

	_, err := h.Read(fds[0])

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseHost(t *testing.T) {
	host, err := parseHost([]byte(
		"GET / HTTP/1.1\r\n" +
			"Host: example.com\r\n",
	))

	if err != nil {
		t.Fatal(err)
	}

	if host != "example.com" {
		t.Fatalf("got %q", host)
	}
}