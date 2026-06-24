package admin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/africanecMorj/mitigation-proxy.git/internal/runtime"

	"github.com/jedib0t/go-pretty/v6/table"
)

const SocketPath = "/tmp/mitigation.sock"

func Reload(path string) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "reload",
		Config:  path,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(resp.Message)

	return nil
}

func Drain(cluster, address string) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "drain",
		Cluster: cluster,
		Backend: address,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(resp.Message)

	return nil
}

func UnDrain(cluster, address string) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "undrain",
		Cluster: cluster,
		Backend: address,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp Response

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	if !resp.OK {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(resp.Message)

	return nil
}

func Stats() error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "stats",
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp map[string][]runtime.BackendStats

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	renderTable(resp)

	return nil
}

func StatsCluster(cluster string) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "stats",
		Cluster: cluster,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp map[string][]runtime.BackendStats

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	renderTable(resp)

	return nil
}

func StatsBackend(cluster, backend string) error {
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	req := Request{
		Command: "stats",
		Cluster: cluster,
		Backend: backend,
	}

	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return err
	}

	var resp map[string][]runtime.BackendStats

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return err
	}

	renderTable(resp)

	return nil
}

func renderTable (resp map[string][]runtime.BackendStats) {
	for cluster, backends := range resp {

		t := table.NewWriter()

		t.SetOutputMirror(os.Stdout)

		t.SetTitle("Cluster: " + cluster)

		t.AppendHeader(table.Row{
			"Backend",
			"State",
			"Active",
			"Requests",
			"Success",
			"Failure",
			"Latency",
			"TTFB",
		})

		for _, b := range backends {
			t.AppendRow(table.Row{
				b.Address,
				StateLabel(b.State),
				b.ActiveConnections,
				b.Requests,
				b.Successes,
				b.Failures,
				b.EWMALatency,
				b.EWMATTFB,
			})
			// fmt.Println(b)

		}

		t.Render()
	}
}

func StateLabel(s string) string {
	switch s {
	case "healthy":
		return "🟢 healthy"
	case "suspect":
		return "🟡 suspect"
	case "recovering":
		return "🔴 recovering"
	case "draining":
		return "🧟‍♂️ draining"
	case "unhealthy":
		return "🔴 unhealthy"
	case "removed":
		return "💀 removed"
	default:
		return s
	}
}
