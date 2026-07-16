# HeavyRain

HeavyRain is a high-performance Layer 4 proxy and load balancer written in Go.

It supports protocol-aware routing, zero-downtime configuration reloads, graceful backend draining, and multiple load balancing algorithms while keeping the data path as lightweight as possible.

---

## Features

- 🚀 High-performance TCP proxy
- 🔒 TLS SNI routing
- 🌐 Wildcard host matching
- 📡 ALPN-based routing (gRPC, HTTP/2, etc.)
- 🐘 PostgreSQL startup packet routing
- ⚖️ Multiple load balancing algorithms
- 🔄 Zero-downtime configuration reload
- 🩺 Backend draining for maintenance
- 📊 Runtime statistics
- ❤️ Passive health monitoring

---
## Building

### Requirements

* Go 1.24+
* Git

Clone the repository:

```bash
git clone https://github.com/africanecMorj/mitigation-proxy.git
cd mitigation-proxy
```

Build the binary:

```bash
go build -o heavyrain ./cmd
```

The executable will be created in the current directory:

```text
./heavyrain
```

Run it:

```bash
./heavyrain
```

---

## Cross-compiling

Build for Linux (AMD64):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o heavyrain-linux-amd64 ./cmd
```

Build for Linux (ARM64):

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o heavyrain-linux-arm64 ./cmd
```

---

## Building release packages

To generate release archives and Linux packages (`.deb`, `.rpm`, `.apk`) install GoReleaser:

```bash
go install github.com/goreleaser/goreleaser/v2@latest
```

Then run:

```bash
goreleaser release --snapshot --clean
```

All generated artifacts will be available in the `dist/` directory.


---

# Commands

Start the proxy:

```bash
heavyrain start config.yaml
```

Reload configuration without dropping active connections:

```bash
heavyrain reload config.yaml
```

Show runtime statistics:

```bash
heavyrain stats
```

Gracefully remove a backend from load balancing:

```bash
heavyrain drain <cluster> <address>
```

Return a backend back into rotation:

```bash
heavyrain undrain <cluster> <address>
```

---

# Configuration

A configuration consists of two primary objects:

- **Listeners** — accept incoming connections and decide where they should be routed.
- **Clusters** — groups of backend servers together with a load balancing policy.

Example:

```yaml
listeners:
  - name: main-tcp
    address: 127.0.0.1:4000

    routing:
      type: tls
      rules:
        - cluster: web
          host: api.example.com

  - name: grpc
    address: 127.0.0.1:4001

    routing:
      type: tls
      rules:
        - cluster: grpc
          alpn:
            - h2

  - name: db
    address: 127.0.0.1:4002

    routing:
      type: postgres
      rules:
        - cluster: postgres
          metadata:
            user: postgres
            password: veryunsualsecretpassword

clusters:
  - name: web
    lb: least_connections
    backends:
      - address: 127.0.0.1:8080
      - address: 127.0.0.1:8081
      - address: 127.0.0.1:8082

  - name: grpc
    lb: p2c
    backends:
      - address: 127.0.0.1:5000
      - address: 127.0.0.1:5001
      - address: 127.0.0.1:5002

  - name: postgres
    lb: least_connections
    backends:
      - address: 127.0.0.1:5432
      - address: 127.0.0.1:5433
```

---

# Routing

HeavyRain supports protocol-aware routing.

## TLS Routing

TLS listeners inspect the ClientHello without terminating TLS.

Routing rules may match on:

- SNI hostname
- ALPN protocol

Example:

```yaml
routing:
  type: tls
  rules:
    - cluster: web
      host: api.example.com

    - cluster: grpc
      alpn:
        - h2
```

No TLS decryption is performed.

---

## Wildcard Hosts

Wildcard matching is supported.

Examples:

```yaml
host: "*.example.com"
```

matches

```
api.example.com
cdn.example.com
foo.example.com
```

while

```yaml
host: "api.example.com"
```

matches only

```
api.example.com
```

More specific rules take precedence over broader wildcard rules.

---

## PostgreSQL Routing

HeavyRain can inspect PostgreSQL startup packets before forwarding the connection.

Routing can be performed using startup parameters such as:

- user
- database
- application_name
- and other startup metadata

Example:

```yaml
routing:
  type: postgres

  rules:
    - cluster: postgres
      metadata:
        user: postgres
```

---

# Load Balancing

Each cluster uses its own balancing strategy.

Currently supported:

## Least Connections

Routes new connections to the backend currently serving the fewest active connections.

Recommended for long-lived TCP connections.

```yaml
lb: least_connections
```

---

## Power of Two Choices (P2C)

Randomly samples two backends and selects the less loaded one.

Provides near-optimal balancing while requiring minimal overhead.

```yaml
lb: p2c
```

## Hash ring

Routes new connections depending on the client ip and backend adress hashes.

Subsequent connections from the same user will be routed to the same backend.

```yaml
lb: sticky
```

---

# Backend Lifecycle

Every backend has a runtime lifecycle.

```
Healthy
   │
   ▼
Draining
   │
   ▼
Removed
```

## Healthy

The backend receives new connections normally.

---

## Draining

A draining backend:

- stops receiving new connections
- continues serving existing connections
- is removed automatically once all active connections finish

Enable draining:

```bash
heavyrain drain web 127.0.0.1:8080
```

Cancel draining:

```bash
heavyrain undrain web 127.0.0.1:8080
```

This allows rolling deployments without interrupting clients.

---

# Zero-Downtime Reload

Reloading configuration does **not** interrupt existing connections.

```bash
heavyrain reload config.yaml
```

During reload HeavyRain:

- reloads listeners
- updates routing rules
- updates clusters
- adds new backends
- removes obsolete backends gracefully
- preserves active connections whenever possible

New connections immediately begin using the updated configuration.

---

# Runtime Statistics

Runtime metrics can be displayed with:

```bash
heavyrain stats
```

Statistics include information such as:

- active connections
- backend state
- bytes transferred
- connection counters
- load balancer status

---

# Prometheus Metrics

HeavyRain exposes runtime metrics in the Prometheus exposition format.

Metrics include:

- Active connections
- Accepted and closed connections
- Bytes sent and received
- Backend request counts
- Backend success and failure counters
- Active backend connections
- Backend state (healthy, draining, unavailable)
- Configuration reloads
- Drain operations

These metrics can be scraped directly by Prometheus and visualized using Grafana, making it easy to monitor traffic, backend health, and load balancing behavior in production.

---

# Design Goals

HeavyRain is designed around a few simple principles:

- minimal latency
- zero-copy data forwarding where possible
- graceful operational workflows
- protocol-aware routing without protocol termination
- efficient load balancing for long-lived TCP services
- safe runtime reconfiguration

---

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the LICENSE file for details.
