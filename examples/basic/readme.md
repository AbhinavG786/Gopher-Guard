# Gopher-Guard: Basic Integration Example

This example demonstrates how to integrate the Gopher-Guard rate limiter into your own Go applications using the official Client SDK.

Gopher-Guard is a **Self-Hosted Infrastructure** tool. You run the cluster inside your own environment (via Docker), and your applications connect to it locally.

## 🛠️ Step 1: Start the Gopher-Guard Engine

You don't need to clone the source code to run Gopher-Guard. You can simply pull the pre-compiled Docker image from Docker Hub.

1. Open your terminal in this directory.
2. Boot the standalone Gopher-Guard node:

```bash
docker compose up -d
```

📝 **Note on Architecture :**
This example uses a lightweight, single-node Gopher-Guard container for quick integration testing and low CPU usage. It will perfectly simulate rate-limiting behavior, but it does not demonstrate Raft consensus

To test the full fault-tolerant, 5-node cluster with Nginx failover and Grafana observability, please use the docker-compose.yml located in the root of the main repository.

## 📦 Step 2: Install the Go SDK

In your own Go project, install the client package:

```
go get github.com/AbhinavG786/Gopher-Guard/client
```

*1. The Example Docker Compose File*

   **File** : examples/basic/docker-compose.yml

```yaml
services:
	gopher-guard:
		image: abhinavg786/gopher-guard:v1.1.0
		ports:
			- "50051:50051" # gRPC Port
			- "8080:8080"   # Admin/Join Port
		environment:
			- PORT=50051
			- ADMIN_PORT=8080
			- NODE_ID=node-1
			- BOOTSTRAP=true
			- RAFT_BIND=0.0.0.0:7000
```

*2. The Example Go Application*

   **File**: examples/basic/main.go

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	// Importing your public SDK
	"github.com/AbhinavG786/Gopher-Guard/client"
)

func main() {
	fmt.Println("🚀 Starting Gopher-Guard Integration Example...")

	// 1. Connect to the local Gopher-Guard container
	// (If using the full production Nginx mesh, this would be localhost:80)
	guard, err := client.New("localhost:50051")
	if err != nil {
		log.Fatalf("Fatal: Could not connect to Gopher-Guard: %v", err)
	}
	defer guard.Close()

	// 2. Define the Rate Limit Rule
	userID := "customer_123"
	limit := 3
	window := 10 * time.Second

	fmt.Printf("Protecting API for %s (Limit: %d requests per %v)\n", userID, limit, window)
	fmt.Println("--------------------------------------------------")

	// 3. Simulate traffic (5 rapid requests)
	for i := 1; i <= 5; i++ {
		// The SDK handles all gRPC serialization and failover retries
		allowed, remaining, err := guard.Allow(context.Background(), userID, limit, window)

		if err != nil {
			log.Printf("[Req %d] ⚠️ Cluster Error: Defaulting to fail-open. Error: %v\n", i, err)
			continue
		}

		if allowed {
			fmt.Printf("[Req %d] ✅ API Call Succeeded | Remaining Quota: %d\n", i, remaining)
		} else {
			fmt.Printf("[Req %d] 🚫 API Call Blocked   | HTTP 429 Too Many Requests\n", i)
		}

		time.Sleep(200 * time.Millisecond)
	}
}
```

## 🚀 Step 3: Run the Example Code
Run the provided main.go file to simulate incoming API traffic:

```bash
go run main.go
```

**Expected Output:**
You will see the first 3 requests succeed, and the subsequent 2 requests instantly blocked by the rate limiter, proving that your API is protected.

🚀 Starting Gopher-Guard Integration Example...
Protecting API for customer_123 (Limit: 3 requests per 10s)

```
[Req 1] ✅ API Call Succeeded | Remaining Quota: 2
[Req 2] ✅ API Call Succeeded | Remaining Quota: 1
[Req 3] ✅ API Call Succeeded | Remaining Quota: 0
[Req 4] 🚫 API Call Blocked | HTTP 429 Too Many Requests
[Req 5] 🚫 API Call Blocked | HTTP 429 Too Many Requests
```

🧹 Cleanup
To stop the Docker container and clear the Raft data:
```
docker compose down -v
```

---

