package grpc

import (
	"context"
	"time"
	"log/slog"
	"github.com/AbhinavG786/Gopher-Guard/internal/engine"
	"github.com/AbhinavG786/Gopher-Guard/internal/grpc/pb"
)

type Server struct{
	pb.UnimplementedRateLimiterServer
	Limiter *engine.SlidingWindow
}

func (s *Server) Check(ctx context.Context, req *pb.RateLimitRequest) (*pb.RateLimitResponse,error){
	slog.Info("Received check request", slog.String("key", req.Key),slog.Int("limit", int(req.Limit)), slog.Int("window_ms", int(req.WindowMs)))

	windowDuration:=time.Duration(req.WindowMs)*time.Millisecond
	allowed,remaining:=s.Limiter.Allow(req.Key,req.Limit,windowDuration)

	return &pb.RateLimitResponse{
		Allowed: allowed,
		Remaining: remaining,
	},nil
}