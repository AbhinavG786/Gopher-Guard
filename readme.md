# 🛡️ Gopher-Guard

A highly available, distributed rate-limiting microservice written in Go. Gopher-Guard implements a strongly consistent, fault-tolerant cluster using the **Raft Consensus Algorithm**, enforcing API rate limits across multiple physical nodes without external dependencies like Redis.

## 🚀 Technical highlights

- **Consensus & High Availability:** Uses `github.com/hashicorp/raft` for leader election and log replication.
- **Contract-First RPC:** gRPC and Protocol Buffers for strict typing and high-performance RPCs.
- **Sliding Window Engine:** Thread-safe sliding-window rate limiter using `sync.RWMutex` to avoid boundary spikes.
- **Embedded Storage:** `bbolt` for fast, embedded persistence of Raft logs and stable state.
- **Background Janitor:** Periodic goroutine that sweeps the in-memory FSM to remove stale entries.

---

## 🏗️ Architecture

```text
[ Client ]
    │ (gRPC / HTTP/2)
    ▼
[ NGINX Load Balancer ] (Port 80)
    │
    ▼ (Round-Robin with Failover)
[ gRPC Server (Leader) ] ──► [ Sliding Window Engine ]
    │                                │
    │                                ▼ (Propose Log)
    └─► [ HashiCorp Raft ] ◄──► [ BoltDB ]
             │
      (Log Replication) ──► [ Node 2, 3, 4, 5 ]
```

## Prerequisites

- Go 1.21+
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins

## 🚦 Running locally with Docker Compose (recommended)

The project now provides a Docker Compose configuration that boots a multi-node cluster (5 nodes by default) plus an NGINX load balancer and observability stack (Prometheus + Grafana). This is the recommended way to start a reproducible local environment.

1. Start the stack

```bash
docker compose up --build -d
```

This will build the image and start the nodes. Containers expose the following useful ports on `localhost`:

- gRPC / API (proxied via NGINX): `80` (the client in `cmd/client` connects to `127.0.0.1:80`)
- Node admin ports: `8080`..`8084` (node-specific admin endpoints)
- Prometheus: `9090`
- Grafana: `3000` (login: `admin` / `admin`)

2. Form the Cluster

By default, Node 1 starts in bootstrap mode. You must manually join the other nodes to form the quorum. Run these from the host (they target the container admin ports):

```bash
curl "http://localhost:8080/join?id=node-2&addr=gopher-node-2:7001"
curl "http://localhost:8081/join?id=node-3&addr=gopher-node-3:7002"
curl "http://localhost:8082/join?id=node-4&addr=gopher-node-4:7003"
curl "http://localhost:8083/join?id=node-5&addr=gopher-node-5:7004"
```

3. Run the gRPC test client (pointed at the load balancer)

```bash
go run cmd/client/main.go
```

3. Stop and remove the stack and volumes

```bash
docker compose down -v
```

4. (Optional) Clean persisted Raft data on the host

```bash
rm -rf ./docker-data
```

Legacy: there is still a `start_cluster.sh` script that can be used for simple local bootstrapping without Docker. The Docker Compose flow is preferred for reproducible, observable environments.

## 🧪 Simulating fault tolerance

To test leader failover with Docker Compose:

1. Identify the leader via container logs (look for election/leadership messages) or check the node admin ports.
2. Simulate an immediate crash (assassination test)

To emulate a sudden crash (SIGKILL) and observe rapid leader re-election, use `docker kill` instead of `docker stop`:

```bash
docker kill gopher-node-1
```

3. Watch the logs of other nodes to observe a new election and leadership change:

```bash
docker logs -f gopher-node-2
```

The cluster should elect a new leader and continue serving requests without data loss.

## 📊 Observability & Dashboards

- **Prometheus:** Scrapes metrics from all nodes every 5s (configured in `prometheus.yml`).
- **Grafana:** Dashboard is pre-provisioned via the repository's provisioning directory and is available at http://localhost:3000.
- **Anonymous Access:** Grafana is configured for anonymous read access so the dashboard is immediately viewable (no manual provisioning required).

Open Grafana to inspect cluster health, per-node request rates, and rate-limit spikes.

## 📦 Dependencies

- `google.golang.org/grpc` — gRPC framework
- `github.com/hashicorp/raft` — Consensus algorithm
- `github.com/hashicorp/raft-boltdb/v2` — Raft storage backend
- `go.etcd.io/bbolt` (bbolt) — embedded key/value store
- `github.com/joho/godotenv` — environment configuration helper

---

For development questions or help running the cluster, open an issue or contact the maintainer.
