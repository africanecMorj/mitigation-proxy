package admin

import (
	"encoding/json"
	"fmt"
	"go.yaml.in/yaml/v4"
	"net"
	"os"
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/runtime"
)

func handle(conn net.Conn, rt *runtime.Runtime) {
	defer conn.Close()

	var req Request

	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return
	}

	switch req.Command {

	case "reload":

		cfg, err := loadConfig(req.Config)
		if err != nil {
			err = fmt.Errorf("Reload error:%w", err)
			json.NewEncoder(conn).Encode(Response{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		err = rt.Reload(cfg)
		if err != nil {
			err = fmt.Errorf("Reload error:%w", err)
			json.NewEncoder(conn).Encode(Response{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		json.NewEncoder(conn).Encode(Response{
			OK:      true,
			Message: "Successfully reloaded",
		})

	case "stats":
		var stats map[string][]runtime.BackendStats

		if req.Cluster == "" {
			stats = rt.Stats()

		} else if req.Cluster != "" && req.Backend == "" {
			stats = rt.StatsCluster(req.Cluster)

		} else if req.Cluster != "" && req.Backend != "" {
			stats = rt.StatsBackend(req.Cluster, req.Backend)
		}

		json.NewEncoder(conn).Encode(stats)
	
	
	case "drain":
		err := rt.Drain(
			req.Cluster,
			req.Backend,
			30*time.Second,
		)
		if err != nil {
			err = fmt.Errorf("Drain error:%w", err)
			json.NewEncoder(conn).Encode(Response{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		json.NewEncoder(conn).Encode(Response{
			OK:      true,
			Message: "Drain started",
		})

	case "undrain":
		err := rt.Undrain(
			req.Cluster,
			req.Backend,
		)
		if err != nil {
			err = fmt.Errorf("Undrain error:%w", err)
			json.NewEncoder(conn).Encode(Response{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		json.NewEncoder(conn).Encode(Response{
			OK:      true,
			Message: "Successfully undrained",
		})
	
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
