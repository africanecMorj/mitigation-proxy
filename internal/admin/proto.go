package admin

type Request struct {
	Command string `json:"command"`
	Config  string `json:"config"`
	Cluster string `json:"cluster,omitempty"`
	Backend string `json:"backend,omitempty"`
}

type Response struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message"`
}

