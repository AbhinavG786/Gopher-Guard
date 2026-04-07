package client

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/AbhinavG786/Gopher-Guard/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.RateLimiterClient
}

func New(target string) (*Client, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gopher-guard: %w", err)
	}
	return &Client{
		conn:       conn,
		grpcClient: pb.NewRateLimiterClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {

	if limit > int(math.MaxInt32) || limit < 0 {
		return false, 0, fmt.Errorf("limit %d is out of bounds for int32", limit)
	}

	windowMs := window.Milliseconds()
	if windowMs > int64(math.MaxInt32) || windowMs < 0 {
		return false, 0, fmt.Errorf("window %v is too large (exceeds max int32 milliseconds)", window)
	}

	maxRetries := 10
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		res, err := c.grpcClient.Check(ctx, &pb.RateLimitRequest{
			Key:      key,
			Limit:    int32(limit),    //nolint:gosec
			WindowMs: int32(windowMs), //nolint:gosec
		})
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}
		return res.Allowed, int(res.Remaining), nil
	}
	return false, 0, fmt.Errorf("failed to check rate limit after %d attempts: %w", maxRetries, lastErr)
}
