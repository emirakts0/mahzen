package handler

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/emirakts0/mahzen/gen/go/mahzen/v1"
	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// entryHandler implements the gRPC EntryServiceServer.
type entryHandler struct {
	pb.UnimplementedEntryServiceServer
	svc *service.EntryService
}

// RegisterEntryServer registers the EntryService gRPC handler.
func RegisterEntryServer(s *grpc.Server, svc *service.EntryService) {
	pb.RegisterEntryServiceServer(s, &entryHandler{svc: svc})
}

func (h *entryHandler) CreateEntry(ctx context.Context, req *pb.CreateEntryRequest) (*pb.CreateEntryResponse, error) {
	userID := userIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	vis := protoToVisibility(req.Visibility)

	entry, err := h.svc.CreateEntry(ctx, userID, req.Title, req.Content, req.Path, vis, req.TagIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating entry: %v", err)
	}

	return &pb.CreateEntryResponse{
		Entry: domainEntryToProto(entry, nil),
	}, nil
}

func (h *entryHandler) GetEntry(ctx context.Context, req *pb.GetEntryRequest) (*pb.GetEntryResponse, error) {
	entry, err := h.svc.GetEntry(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "entry not found: %v", err)
	}

	return &pb.GetEntryResponse{
		Entry: domainEntryToProto(entry, nil),
	}, nil
}

func (h *entryHandler) UpdateEntry(ctx context.Context, req *pb.UpdateEntryRequest) (*pb.UpdateEntryResponse, error) {
	vis := protoToVisibility(req.Visibility)

	entry, err := h.svc.UpdateEntry(ctx, req.Id, req.Title, req.Content, req.Path, vis, req.TagIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "updating entry: %v", err)
	}

	return &pb.UpdateEntryResponse{
		Entry: domainEntryToProto(entry, nil),
	}, nil
}

func (h *entryHandler) DeleteEntry(ctx context.Context, req *pb.DeleteEntryRequest) (*pb.DeleteEntryResponse, error) {
	if err := h.svc.DeleteEntry(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "deleting entry: %v", err)
	}

	return &pb.DeleteEntryResponse{}, nil
}

func (h *entryHandler) ListEntries(ctx context.Context, req *pb.ListEntriesRequest) (*pb.ListEntriesResponse, error) {
	userID := userIDFromContext(ctx)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	entries, total, err := h.svc.ListEntries(ctx, userID, req.Path, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing entries: %v", err)
	}

	pbEntries := make([]*pb.Entry, len(entries))
	for i, e := range entries {
		pbEntries[i] = domainEntryToProto(e, nil)
	}

	return &pb.ListEntriesResponse{
		Entries: pbEntries,
		Total:   int32(total),
	}, nil
}

// domainEntryToProto converts a domain Entry to a protobuf Entry.
func domainEntryToProto(e *domain.Entry, tags []string) *pb.Entry {
	pbEntry := &pb.Entry{
		Id:         e.ID,
		UserId:     e.UserID,
		Title:      e.Title,
		Content:    e.Content,
		Summary:    e.Summary,
		Visibility: visibilityToProto(e.Visibility),
		Tags:       tags,
		CreatedAt:  timestamppb.New(e.CreatedAt),
		UpdatedAt:  timestamppb.New(e.UpdatedAt),
		Path:       e.Path,
	}
	return pbEntry
}

func protoToVisibility(v pb.Visibility) domain.Visibility {
	switch v {
	case pb.Visibility_VISIBILITY_PUBLIC:
		return domain.VisibilityPublic
	case pb.Visibility_VISIBILITY_PRIVATE:
		return domain.VisibilityPrivate
	default:
		return domain.VisibilityPrivate
	}
}

func visibilityToProto(v domain.Visibility) pb.Visibility {
	switch v {
	case domain.VisibilityPublic:
		return pb.Visibility_VISIBILITY_PUBLIC
	case domain.VisibilityPrivate:
		return pb.Visibility_VISIBILITY_PRIVATE
	default:
		return pb.Visibility_VISIBILITY_UNSPECIFIED
	}
}
