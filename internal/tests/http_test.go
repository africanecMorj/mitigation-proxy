package tests

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"io"
	"strings"
	"time"

    "github.com/africanecMorj/mitigation-proxy.git/internal/runtime"
    "github.com/africanecMorj/mitigation-proxy.git/internal/config"
)


func BenchmarkHTTPRequest(b *testing.B) {
	backend := startBackend()
	defer backend.Close()

		cfg := &config.Config{
		Listeners: []config.Listener{
			{
				Name:    "http",
				Address: "127.0.0.1:8080",

				Routing: config.Routing{
					Type:           "http",
					DefaultCluster: "backend",

					Rules: []config.Rule{
						{
							Name:    "localhost",
							Host:    "localhost",
							Cluster: "backend",
						},
					},
				},
			},
		},

		Clusters: []config.Cluster{
			{
				Name: "backend",
				LB:   "round_robin",

				Backends: []config.Backend{
					{
						Address: strings.TrimPrefix(backend.URL, "http://"),
						Weight:  1,
						Tau:     0.5,
					},
				},
			},
		},
	}

    rt := runtime.New()
	err := rt.Build(cfg)
	if err != nil {
		b.Fatal(err)
	}

	time.Sleep(10*time.Second)

	defer func (){
		b.StopTimer()
		rt.Shutdown(1*time.Second, nil)
	}()


	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	req, err := http.NewRequest(
		"GET",
		"http://127.0.0.1:8080/",
		nil,
	)
	if err != nil {
		b.Fatal(err)
	}

	req.Host = "localhost"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}

		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func startBackend() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			},
		),
	)
}