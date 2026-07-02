package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.yaml.in/yaml/v4"

	"github.com/africanecMorj/mitigation-proxy.git/internal/admin"
	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/runtime"
	"github.com/africanecMorj/mitigation-proxy.git/internal/metrics/prometheus"
	"github.com/africanecMorj/mitigation-proxy.git/internal/metrics"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: heavyrain <start|reload|stats> <config.yaml>")
	}

	switch os.Args[1] {

	case "start":
		if len(os.Args) < 3 {
			log.Fatal("usage: heavyrain start <config.yaml>")
		}
		
		start(os.Args[2])

	case "reload":
		if len(os.Args) < 3 {
			log.Fatal("usage: heavyrain reload <config.yaml>")
		}
		if err := admin.Reload(os.Args[2]); err != nil {
			log.Fatal(err)
		}
	case "stats":
		if len(os.Args) == 2 {
			if err := admin.Stats(); err != nil {
				log.Fatal(err)
			}
		} else if len(os.Args) == 3 {
			if err := admin.StatsCluster(os.Args[2]); err != nil {
				log.Fatal(err)
			}
		
		} else if len(os.Args) > 3 {
			if err := admin.StatsBackend(os.Args[2], os.Args[3]); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("usage: heavyrain stats <cluster> <backend>")
		}

	case "drain":
		if len(os.Args) < 4 {
			log.Fatal("usage: heavyrain drain <cluster> <backend address>")
		}
		if err := admin.Drain(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	case "undrain":
		if len(os.Args) < 4 {
			log.Fatal("usage: heavyrain drain <cluster> <backend address>")
		}
		if err := admin.UnDrain(os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

func start(configPath string) error {

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	rt := runtime.New()

	adminServer, err := admin.Start(rt)
	if err != nil {
		log.Fatal(err)
	}
	defer adminServer.Close()

	var metrics metrics.Server
	if cfg.Global.Prometheus {
		metrics, err = prometheus.Start(":9090", rt)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		metrics = nil
	}
	

	if err := rt.Build(cfg, metrics); err != nil {
		log.Fatal(err)
	}

	waitForShutdown()

	err = rt.Shutdown(rt.ShutdownTimeout(), metrics)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func waitForShutdown() {

	sigCh := make(chan os.Signal, 1)

	signal.Notify(
		sigCh,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	sig := <-sigCh

	log.Printf(
		"received signal: %s",
		sig,
	)
}

func loadConfig(path string) (*config.Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
