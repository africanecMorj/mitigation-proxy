package config

import "time"

type Config struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
	Global    Global     `yaml:"global"`
}

type Listener struct {
	Name     string  `yaml:"name"`
	Address  string  `yaml:"address"`
	Protocol string  `yaml:"protocol"`
	Routing  Routing `yaml:"routing"`
}

type Routing struct {
	Type   string `yaml:"type"`
	Rules  []Rule `yaml:"rules"`

	DefaultCluster string `yaml:"default_cluster"`
}

type Rule struct {
	Name    string   `yaml:"name"`
	Host    string   `yaml:"host"`
	ALPN    []string `yaml:"alpn"`
	Cluster string   `yaml:"cluster"`
	Default bool `yaml:"default"`
}

type Cluster struct {
	Name     string    `yaml:"name"`
	LB       string    `yaml:"lb"`
	Backends []Backend `yaml:"backends"`
}

type Backend struct {
	Address string `yaml:"address"`
	Weight  int64    `yaml:"weight"`
	Tau  float64    `yaml:"tau"`
}

type Global struct {
	Workers int `yaml:"workers"`

	Limits struct {
		MaxConnections int `yaml:"max_connections"`
	} `yaml:"limits"`

	Timeouts struct {
		ClientHello time.Duration `yaml:"client_hello"`
		Idle        time.Duration `yaml:"idle"`
	} `yaml:"timeouts"`
}