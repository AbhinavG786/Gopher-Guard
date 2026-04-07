package main

import (
	"context"
	"github.com/AbhinavG786/Gopher-Guard/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	conn, err := grpc.NewClient("127.0.0.1:80", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewRateLimiterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Sending check request for user_123")
	for i := 1; i <= 5; i++ {
		time.Sleep(100 * time.Millisecond)
		maxRetries := 10
		var success bool
		for attempt := 1; attempt <= maxRetries; attempt++ {
			res, err := client.Check(ctx, &pb.RateLimitRequest{
				Key:      "user_123",
				Limit:    3,
				WindowMs: 60000,
			})
			if err != nil {
				log.Printf("[Req %d] Attempt %d failed: %v. Retrying...", i, attempt, err)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			log.Printf("[Req %d] Allowed: %v, Remaining: %d", i, res.Allowed, res.Remaining)
			success = true
			break
		}
		if !success {
			log.Fatalf("[Req %d] Completely failed after %d retries", i, maxRetries)
		}
	}
}
