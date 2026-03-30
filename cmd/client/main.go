package main

import (
	"context"
	"log"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
)

func main(){
	conn,err:= grpc.NewClient("localhost:50051",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err!=nil{
		log.Fatalf("Failed to connect: %v",err)
	}
	defer conn.Close()
	client:=pb.NewRateLimiterClient(conn)
	ctx,cancel:=context.WithTimeout(context.Background(),time.Second)
	defer cancel()

	log.Println("Sending check request for user_123")
	for range 5{
		time.Sleep(100 * time.Millisecond)
		res,err:=client.Check(ctx,&pb.RateLimitRequest{
			Key: "user_123",
			Limit: 3,
			WindowMs: 60000,
		})
	if err!=nil{
		log.Fatalf("Failed to check rate limit: %v",err)
	}
	log.Printf("Allowed: %v, Remaining: %d",res.Allowed,res.Remaining)
}
}