package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/emirakts0/mahzen/gen/go/mahzen/v1"
	"github.com/emirakts0/mahzen/internal/service"
)

// searchHandler implements the gRPC SearchServiceServer.
type searchHandler struct {
	pb.UnimplementedSearchServiceServer
	svc *service.SearchService
}

// RegisterSearchServer registers the SearchService gRPC handler.
func RegisterSearchServer(s *grpc.Server, svc *service.SearchService) {
	pb.RegisterSearchServiceServer(s, &searchHandler{svc: svc})
}

func (h *searchHandler) KeywordSearch(ctx context.Context, req *pb.KeywordSearchRequest) (*pb.SearchResponse, error) {
	userID := userIDFromContext(ctx)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	results, total, err := h.svc.KeywordSearch(ctx, req.Query, userID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "keyword search: %v", err)
	}

	pbResults := make([]*pb.SearchResult, len(results))
	for i, r := range results {
		pbResults[i] = &pb.SearchResult{
			EntryId:    r.EntryID,
			Title:      r.Title,
			Snippet:    r.Snippet,
			Score:      r.Score,
			Highlights: r.Highlights,
		}
	}

	return &pb.SearchResponse{
		Results: pbResults,
		Total:   int32(total),
	}, nil
}

func (h *searchHandler) SemanticSearch(ctx context.Context, req *pb.SemanticSearchRequest) (*pb.SearchResponse, error) {
	userID := userIDFromContext(ctx)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	results, total, err := h.svc.SemanticSearch(ctx, req.Query, userID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "semantic search: %v", err)
	}

	pbResults := make([]*pb.SearchResult, len(results))
	for i, r := range results {
		pbResults[i] = &pb.SearchResult{
			EntryId:    r.EntryID,
			Title:      r.Title,
			Snippet:    r.Snippet,
			Score:      r.Score,
			Highlights: r.Highlights,
		}
	}

	return &pb.SearchResponse{
		Results: pbResults,
		Total:   int32(total),
	}, nil
}
