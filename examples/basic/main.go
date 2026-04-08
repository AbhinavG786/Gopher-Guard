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

	// (If using the full production Nginx mesh, this would be localhost:80)
	// Connect to the Gopher-Guard Nginx Load Balancer assuming you are running this locally via Docker Compose
	// guard, err := client.New("localhost:80")
	
	// 1. Connect to the local Gopher-Guard container
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