package rpcfiles

import (
	"context"
	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"time"
)

// Fetch asks the node at address for its file metadata. The RPC does the real network I/O;
// creating the client connection object does NOT prove the server is reachable.
func Fetch(ctx context.Context, address, requester string, timeout time.Duration) (*pb.ListFilesResponse, error) {
	if !config.ValidAddress(address) {
		return nil, status.Error(codes.InvalidArgument, "invalid or non-loopback gRPC address")
	}
	if timeout <= 0 {
		return nil, status.Error(codes.InvalidArgument, "timeout must be positive")
	}
	opCtx, cancel := context.WithTimeout(ctx, timeout) // bound the call; also honours caller cancellation
	defer cancel()

	conn, err := grpc.NewClient(
		address,
		// Unencrypted ONLY because CW1 is a closed loopback teaching setup.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(4096),
			grpc.MaxCallRecvMsgSize(65536),
		),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "create gRPC client: "+err.Error())
	}
	defer conn.Close()

	resp, err := pb.NewFileServiceClient(conn).ListFiles(opCtx, &pb.ListFilesRequest{RequesterId: requester})
	if err != nil {
		return nil, err // keep the structured status; never turn a failure into an empty list
	}
	return resp, nil
}