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
[ gRPC Server ] ──► [ Sliding Window Engine ]
                        │
                        ▼ (Propose Log)
[ HashiCorp Raft (Leader) ] ◄──► [ BoltDB ] (Log persistence)
    │
    ├─► Replicate to Node 2
    └─► Replicate to Node 3
            │
            ▼ (Quorum reached)
[ Limiter FSM (Apply Commit) ] ──► [ In-memory map updated ]
```

## Prerequisites

- Go 1.21+
- `protoc` (Protocol Buffers compiler)
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins

## 🚦 Running a local 3-node cluster

Gopher-Guard includes a script to bootstrap a 3-node cluster locally for development.

1. Clone and start the cluster

```bash
git clone https://github.com/AbhinavG786/Gopher-Guard.git
cd Gopher-Guard
chmod +x start_cluster.sh
./start_cluster.sh
```

Node 1 will bootstrap as the leader; Node 2 and Node 3 will wait to join.

2. Join the other nodes (in separate terminals)

```bash
curl "http://localhost:8080/join?id=node-2&addr=127.0.0.1:7001"
curl "http://localhost:8080/join?id=node-3&addr=127.0.0.1:7002"
```

3. Run the gRPC test client

```bash
go run cmd/client/main.go
```

You should observe requests being proposed to the Raft leader, replicated to followers, applied to the FSM, and returned with quota information.

## 🧪 Simulating fault tolerance

To test leader failover:

1. Find the PID of the leader process (listening on the gRPC port).
2. Kill it (e.g., `kill -9 <PID>`).
3. Watch the logs for a new election; a follower will become leader and the cluster continues serving requests without data loss.

## 📦 Dependencies

- `google.golang.org/grpc` — gRPC framework
- `github.com/hashicorp/raft` — Consensus algorithm
- `github.com/hashicorp/raft-boltdb/v2` — Raft storage backend
- `go.etcd.io/bbolt` (bbolt) — embedded key/value store
- `github.com/joho/godotenv` — environment configuration helper

---

For development questions or help running the cluster, open an issue or contact the maintainer.
