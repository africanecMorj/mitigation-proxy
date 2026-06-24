# Heavyrain Mitigation Proxy

High-performance mitigation proxy for TCP, QUIC, SNI/ALPN routing, backend balancing, traffic control, and runtime backend management.

## Features

* TCP proxy
* QUIC / UDP proxy
* SNI + ALPN based routing
* HTTP/2, HTTP/1.1, HTTP/3 support
* Multiple load balancing algorithms:

  * round_robin
  * ewma
  * p2c
* Backend draining and enabling
* Connection limits
* Per-IP rate limiting
* Runtime configuration reload
* Metrics and statistics

---

## Configuration

Example `config.yaml`:

```yaml
listeners:
  - name: main-tcp
    address: ":4000"
    protocol: tcp

    routing:
      type: sni_alpn

      rules:
        - host: api.example.com
          alpn: ["h2"]
          cluster: grpc_cluster

        - host: api.example.com
          alpn: ["http/1.1"]
          cluster: rest_cluster

        - host: "*.example.com"
          cluster: web_cluster

        - default: true
          cluster: default_cluster

  - name: quic
    address: ":4433"
    protocol: udp

    routing:
      type: quic
      default_cluster: h3_cluster


clusters:

  - name: grpc_cluster
    lb: ewma

    pool:

    backends:
      - address: 10.0.0.1:50051
        weight: 1

      - address: 10.0.0.2:50051
        weight: 1


  - name: rest_cluster
    lb: ewma

    pool:
      max_idle: 256

    backends:
      - address: 127.0.0.1:8080

      - address: 127.0.0.1:8080


  - name: web_cluster
    lb: round_robin

    backends:
      - address: 127.0.0.1:8080

      - address: 127.0.0.1:8080


  - name: h3_cluster
    lb: p2c

    backends:
      - address: 127.0.0.1:8080

      - address: 127.0.0.1:8080


  - name: default_cluster
    lb: round_robin

    backends:
      - address: 127.0.0.1:8080


limits:
  max_connections: 100000


ratelimit:
  per_ip_rps: 100
```

---

# Routing

## SNI / ALPN routing

Traffic can be routed based on TLS SNI hostname and ALPN protocol.

Example:

```
api.example.com + h2
        |
        v
   grpc_cluster


api.example.com + http/1.1
        |
        v
   rest_cluster
```

Wildcard routing:

```yaml
- host: "*.example.com"
  cluster: web_cluster
```

Default route:

```yaml
- default: true
  cluster: default_cluster
```

---

# Load Balancing

## Round Robin

```yaml
lb: round_robin
```

Sequential backend selection.

---

## EWMA

```yaml
lb: ewma
```

Latency-aware balancing algorithm.

Recommended for:

* gRPC services
* APIs
* variable latency backends

---

## P2C

```yaml
lb: p2c
```

Power of Two Choices balancing.

Recommended for:

* large backend pools
* uneven traffic distribution

---

# Commands

## Start

Start proxy with configuration:

```bash
heavyrain start config.yaml
```

---

## Reload

Reload configuration without restarting:

```bash
heavyrain reload config.yaml
```

---

## Statistics

Show runtime statistics:

```bash
heavyrain stats
```

---

# Backend Management

## List backends

```bash
mitigation backend list
```

Example:

```
cluster          backend        state

grpc_cluster     api-1          enabled
grpc_cluster     api-2          enabled

rest_cluster     rest-1         draining
```

---

## Drain backend

Remove backend from new traffic while keeping existing connections:

```bash
mitigation backend drain api-1
```

Useful for:

* deployments
* maintenance
* graceful shutdown

---

## Enable backend

Enable backend traffic:

```bash
mitigation backend enable api-1
```

---

# Metrics

Show runtime metrics:

```bash
mitigation metrics
```

Available metrics include:

```
connections_active

requests_total

backend_errors_total

backend_latency

rate_limit_dropped
```

---

# Limits

Maximum connections:

```yaml
limits:
  max_connections: 100000
```

Per-IP rate limit:

```yaml
ratelimit:
  per_ip_rps: 100
```

---

# Runtime Flow

```
                Client

                  |

                  v

          Heavyrain Proxy

                  |

        +---------+---------+

        |                   |

     TCP Listener       QUIC Listener

        |                   |

    SNI / ALPN          HTTP/3

        |

   Cluster Selection

        |

   Backend Load Balancer

        |

     Backend Pool
```

---

# Graceful Deployment

Recommended backend replacement flow:

1. Add new backend

2. Enable backend:

```bash
mitigation backend enable api-new
```

3. Drain old backend:

```bash
mitigation backend drain api-old
```

4. Wait for active connections to finish

---

# License

MIT
