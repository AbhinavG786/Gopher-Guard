package grpc

import (
	"context"
	"fmt"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
)

type Server struct{
	pb.UnimplementedRateLimiterServer
}

func (s *Server) Check(ctx context.Context, req *pb.RateLimitRequest) (*pb.RateLimitResponse,error){
	fmt.Printf("Received check request for key: %s\n", req.Key)

	return &pb.RateLimitResponse{
		Allowed: true,
		Remaining: 99,
	},nil
}