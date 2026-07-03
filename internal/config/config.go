package config

type Config struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
	Global    Global     `yaml:"global"`
}

type Listener struct {
	Address  string  `yaml:"address"`
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
	
	Metadata map[string]string `yaml:"Metadata"`
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
	Prometheus bool `yaml:"prometheus"`
	Dev	 	   bool `yaml:"dev"`

	Limits struct {
		MaxConnections int32 `yaml:"max_connections"`
		Burst int32 		 `yaml:"burst"`
		RPS int32   		 `yaml:"per_ip_rps"`

	} `yaml:"limits"`

	Timeouts struct {
		DrainTimeout string `yaml:"drain"`
		ShutdownTimeout string `yaml:"shutdown"`
		HealthCheckTimeout string `yaml:"healthcheck"`
		
	} `yaml:"timeouts"`
}