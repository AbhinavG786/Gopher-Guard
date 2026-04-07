package grpc

import (
	"context"
	"github.com/AbhinavG786/Gopher-Guard/internal/engine"
	"github.com/AbhinavG786/Gopher-Guard/pkg/pb"
	"log/slog"
	"time"
)

type Server struct {
	pb.UnimplementedRateLimiterServer
	Limiter *engine.SlidingWindow
}

func (s *Server) Check(ctx context.Context, req *pb.RateLimitRequest) (*pb.RateLimitResponse, error) {
	slog.Info("🚨Received check request", slog.String("key", req.Key), slog.Int("limit", int(req.Limit)), slog.Int("window_ms", int(req.WindowMs)))

	windowDuration := time.Duration(req.WindowMs) * time.Millisecond
	allowed, remaining, err := s.Limiter.Allow(req.Key, req.Limit, windowDuration)
	if err != nil {
		// If we aren't the leader, or Raft failed, tell the client
		slog.Error("Raft rejection", slog.String("error", err.Error()))
		return nil, err
	}

	return &pb.RateLimitResponse{
		Allowed:   allowed,
		Remaining: remaining,
	}, nil
}
