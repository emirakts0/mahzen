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

// tagHandler implements the gRPC TagServiceServer.
type tagHandler struct {
	pb.UnimplementedTagServiceServer
	svc *service.TagService
}

// RegisterTagServer registers the TagService gRPC handler.
func RegisterTagServer(s *grpc.Server, svc *service.TagService) {
	pb.RegisterTagServiceServer(s, &tagHandler{svc: svc})
}

func (h *tagHandler) CreateTag(ctx context.Context, req *pb.CreateTagRequest) (*pb.CreateTagResponse, error) {
	tag, err := h.svc.CreateTag(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "creating tag: %v", err)
	}

	return &pb.CreateTagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *tagHandler) GetTag(ctx context.Context, req *pb.GetTagRequest) (*pb.GetTagResponse, error) {
	tag, err := h.svc.GetTag(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tag not found: %v", err)
	}

	return &pb.GetTagResponse{
		Tag: domainTagToProto(tag),
	}, nil
}

func (h *tagHandler) ListTags(ctx context.Context, req *pb.ListTagsRequest) (*pb.ListTagsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)

	tags, total, err := h.svc.ListTags(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "listing tags: %v", err)
	}

	pbTags := make([]*pb.Tag, len(tags))
	for i, t := range tags {
		pbTags[i] = domainTagToProto(t)
	}

	return &pb.ListTagsResponse{
		Tags:  pbTags,
		Total: int32(total),
	}, nil
}

func (h *tagHandler) DeleteTag(ctx context.Context, req *pb.DeleteTagRequest) (*pb.DeleteTagResponse, error) {
	if err := h.svc.DeleteTag(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "deleting tag: %v", err)
	}
	return &pb.DeleteTagResponse{}, nil
}

func (h *tagHandler) AttachTag(ctx context.Context, req *pb.AttachTagRequest) (*pb.AttachTagResponse, error) {
	if err := h.svc.AttachTag(ctx, req.EntryId, req.TagId); err != nil {
		return nil, status.Errorf(codes.Internal, "attaching tag: %v", err)
	}
	return &pb.AttachTagResponse{}, nil
}

func (h *tagHandler) DetachTag(ctx context.Context, req *pb.DetachTagRequest) (*pb.DetachTagResponse, error) {
	if err := h.svc.DetachTag(ctx, req.EntryId, req.TagId); err != nil {
		return nil, status.Errorf(codes.Internal, "detaching tag: %v", err)
	}
	return &pb.DetachTagResponse{}, nil
}

func domainTagToProto(t *domain.Tag) *pb.Tag {
	return &pb.Tag{
		Id:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: timestamppb.New(t.CreatedAt),
	}
}
