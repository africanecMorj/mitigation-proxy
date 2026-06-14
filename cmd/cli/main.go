package main

import (
    "log"
    "os"
    
    "go.yaml.in/yaml/v4"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/cmd/buildLoop"
)

func main() {
    if len(os.Args) != 3 {
        log.Fatal("usage: startProxy start [config.yaml]")
    }

    switch os.Args[1] {
    case "start":

        configPath := os.Args[2]

        cfg, err := loadConfig(configPath)
        if err != nil {
            log.Fatal(err)
        }

        err = buildLoop.Build(cfg)
        if err != nil {
            log.Fatal(err)
        }

    default:
        log.Fatalf("unknown command: %s", os.Args[1])
    }
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