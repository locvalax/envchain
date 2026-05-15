package source

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// GRPCClient defines the interface for fetching secrets over gRPC.
type GRPCClient interface {
	GetSecret(ctx context.Context, key string) (string, error)
	ListKeys(ctx context.Context) ([]string, error)
}

type grpcSource struct {
	client GRPCClient
	prefix string
}

// GRPCClientConfig holds configuration for connecting to a gRPC secrets service.
type GRPCClientConfig struct {
	Address string
	Prefix  string
}

// defaultGRPCClient wraps a real gRPC connection implementing GRPCClient via
// a simple key/value service convention. Replace with your generated proto client.
type defaultGRPCClient struct {
	conn *grpc.ClientConn
}

func (d *defaultGRPCClient) GetSecret(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("not implemented: use a generated proto client")
}

func (d *defaultGRPCClient) ListKeys(_ context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented: use a generated proto client")
}

// NewGRPCSource creates a Source backed by a gRPC secrets service.
func NewGRPCSource(cfg GRPCClientConfig) (Source, error) {
	conn, err := grpc.NewClient(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc_source: dial %s: %w", cfg.Address, err)
	}
	return NewGRPCSourceWithClient(&defaultGRPCClient{conn: conn}, cfg.Prefix), nil
}

// NewGRPCSourceWithClient creates a Source using the provided GRPCClient (useful for testing).
func NewGRPCSourceWithClient(client GRPCClient, prefix string) Source {
	return &grpcSource{client: client, prefix: prefix}
}

func (g *grpcSource) Get(key string) (string, bool) {
	lookup := key
	if g.prefix != "" {
		lookup = g.prefix + key
	}
	val, err := g.client.GetSecret(context.Background(), lookup)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return "", false
		}
		return "", false
	}
	return val, true
}

func (g *grpcSource) Keys() []string {
	keys, err := g.client.ListKeys(context.Background())
	if err != nil {
		return nil
	}
	return keys
}

func (g *grpcSource) Name() string {
	return "grpc"
}
