//First prototype (funny thing)


// package TCPLayer

// import (
// 	"log"
// 	"net"
// 	"sync"
// 	"time"

// 	"github.com/africanecMorj/mitigation-proxy.git/internal/ratelimit"
// 	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
// 	"github.com/africanecMorj/mitigation-proxy.git/pkg"
// 	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers/health"

// )

// const LISTEN_ADDR = ":9000"

// func proxy(
// 	dst net.Conn,
// 	src net.Conn,
// 	wg *sync.WaitGroup,
// ) {
// 	defer wg.Done()

// 	srcTCP, ok := src.(*net.TCPConn)
// 	if !ok {
// 		return
// 	}

// 	dstTCP, ok := dst.(*net.TCPConn)
// 	if !ok {
// 		return
// 	}

// 	srcFile, err := srcTCP.File()
// 	if err != nil {
// 		return
// 	}
// 	defer srcFile.Close()

// 	dstFile, err := dstTCP.File()
// 	if err != nil {
// 		return
// 	}
// 	defer dstFile.Close()

// 	_ = pkg.Splice(
// 		srcFile,
// 		dstFile,
// 	)

// 	_ = dstTCP.CloseWrite()
// 	_ = srcTCP.CloseRead()
// }

// func handleConnection(
// 	b *health.Backend,
// 	clientConn net.Conn,
// ) {

// 	defer clientConn.Close()

// 	b.ActiveConnections.Add(1)
// 	defer b.ActiveConnections.Add(-1)

// 	defer ratelimit.DecrementConnections()

// 	clientAddr := clientConn.RemoteAddr().String()

// 	host, _, err := net.SplitHostPort(clientAddr)
// 	if err != nil {
// 		clientConn.Close()
// 		return
// 	}

// 	limiter := ratelimit.GetLimiter(host)

// 	if !limiter.Allow() {
// 		log.Printf("Rate limited: %s", host)
// 		clientConn.Close()
// 		return
// 	}

// 	start := time.Now()
// 	backendConn, err := net.DialTimeout(
// 		"tcp",
// 		b.Address,
// 		5*time.Second,
// 	)
// 	latency := time.Since(start)

// 	if err != nil {
// 		b.MarkFailure()
// 		log.Printf(
// 			"backend %s unavailable: %v",
// 			b.Address,
// 			err,
// 		)

// 		return
// 	}
// 	b.MarkSuccess(latency)

// 	defer backendConn.Close()

// 	var wg sync.WaitGroup

// 	wg.Add(2)

// 	go proxy(
// 		backendConn,
// 		clientConn,
// 		&wg,
// 	)

// 	go proxy(
// 		clientConn,
// 		backendConn,
// 		&wg,
// 	)

// 	wg.Wait()
// }

// func Init (LISTEN_ADDR string, bl balancers.Balancer){

// 	bl.StartHealthChecks()

// 	listener, err := net.Listen(
// 		"tcp",
// 		LISTEN_ADDR,
// 	)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer listener.Close()

// 	ratelimit.Init()

// 	log.Printf(
// 		"proxy listening on %s",
// 		LISTEN_ADDR,
// 	)

// 	for {

// 		conn, err := listener.Accept()

// 		if err != nil {
// 			continue
// 		}

// 		if !ratelimit.IncrementConnections() {
// 			log.Printf("Connection limit reached")
// 			conn.Close()
// 			continue
// 		}

// 		backend := bl.Next()

// 		if backend == nil {

// 			log.Println(
// 				"no healthy backends",
// 			)

// 			_ = conn.Close()

// 			continue
// 		}

// 		go handleConnection(
// 			backend,
// 			conn,
// 		)
// 	}
// }

