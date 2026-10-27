package rpcfiles

import (
	"context"
	pb "example.edu/6cc545/cw1/api/files/v1"
	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Catalog interface{ Snapshot() []model.FileMetadata }
type Service struct {
	pb.UnimplementedFileServiceServer
	NodeID  string
	Catalog Catalog
}

// ListFiles returns this node's catalogue metadata. No filesystem access, no file contents.
func (s *Service) ListFiles(ctx context.Context, r *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err() // Canceled / DeadlineExceeded
	}
	if r == nil || !config.ValidID(r.GetRequesterId()) {
		return nil, status.Error(codes.InvalidArgument, "requester_id must be a valid node ID")
	}
	if s.Catalog == nil {
		return nil, status.Error(codes.Internal, "catalogue not configured")
	}
	snap := s.Catalog.Snapshot() // the SAME catalogue object the REST /v1/files uses
	files := make([]*pb.FileMetadata, 0, len(snap))
	for _, f := range snap {
		files = append(files, &pb.FileMetadata{
			Name:         f.Name,
			Size:         f.Size,
			Sha256:       f.SHA256,
			ModifiedUnix: f.ModifiedUnix,
		})
	}
	return &pb.ListFilesResponse{Files: files, NodeId: s.NodeID}, nil
}