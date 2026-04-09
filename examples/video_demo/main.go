package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AbhinavG786/Gopher-Guard/client"
)

func main() {
	fmt.Println("🚀 Initializing Gopher-Guard Chaos Test...")

	guard, err := client.New("localhost:80")
	if err != nil {
		panic(err)
	}
	defer guard.Close()

	userID := "demo_user_777"
	limit := 10               
	window := 5 * time.Second 

	fmt.Printf("🛡️ Protecting API for %s (Limit: %d req / %v)\n", userID, limit, window)
	fmt.Println("--------------------------------------------------")

	burstCount := 1

	for {
		fmt.Printf("\n🌊 --- INCOMING TRAFFIC SPIKE %d --- 🌊\n", burstCount)

		for i := 1; i <= 15; i++ {
			allowed, remaining, err := guard.Allow(context.Background(), userID, limit, window)

			if err != nil {
				
				fmt.Printf("[Req %02d] ⚠️ Node is follower. Hunting for new Leader...\n", i)
				
				guard.Close() 
				guard, _ = client.New("localhost:80") 
				
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if allowed {
				fmt.Printf("[Req %02d] ✅ ACCEPTED | Remaining Quota: %d\n", i, remaining)
			} else {
				fmt.Printf("[Req %02d] 🚫 BLOCKED  | Rate Limit Enforced!\n", i)
			}

			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println("\n⏳ Traffic subsided. Waiting for the sliding window to clear (5s)...")
		time.Sleep(5 * time.Second)
		burstCount++
	}
}