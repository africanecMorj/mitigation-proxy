package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"

    "github.com/africanecMorj/mitigation-proxy.git/pkg/splice"
    "github.com/africanecMorj/mitigation-proxy.git/internal/limiter"

	"golang.org/x/time/rate"
)

func proxy(src net.Conn, dst net.Conn) {
	srcFile, err := src.File()
	if err != nil {
    	return err
	}
	defer srcFile.Close()

	dstFile, err := dst.File()
	if err != nil {
		return err
	}
	defer dstFile.Close()

	pkg.Splice(dstFile, srcFile)
}

func handleConnection(clientConn net.Conn) {
	defer decrementConnections()

	clientAddr := clientConn.RemoteAddr().String()

	host, _, err := net.SplitHostPort(clientAddr)
	if err != nil {
		clientConn.Close()
		return
	}

	limiter := getLimiter(host)

	if !limiter.Allow() {
		log.Printf("Rate limited: %s", host)
		clientConn.Close()
		return
	}

	backendConn, err := net.DialTimeout(
		"tcp",
		BACKEND_ADDR,
		5*time.Second,
	)

	if err != nil {
		log.Printf("Backend unavailable: %v", err)
		clientConn.Close()
		return
	}

	clientConn.SetDeadline(time.Now().Add(5 * time.Minute))
	backendConn.SetDeadline(time.Now().Add(5 * time.Minute))

	
	go proxy(clientConn, backendConn)
	go proxy(backendConn, clientConn)
}

func main() {
	listener, err := net.Listen("tcp", LISTEN_ADDR)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Proxy listening on %s", LISTEN_ADDR)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		if !incrementConnections() {
			log.Printf("Connection limit reached")
			conn.Close()
			continue
		}

		go handleConnection(conn)
	}
}