package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/emirakts0/mahzen/gen/go/mahzen/v1"
	"github.com/emirakts0/mahzen/internal/domain"
)

// ---------------------------------------------------------------------------
// domainEntryToProto
// ---------------------------------------------------------------------------

func TestDomainEntryToProto(t *testing.T) {
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	entry := &domain.Entry{
		ID:         "entry-1",
		UserID:     "user-1",
		Title:      "Test Entry",
		Content:    "some content",
		Summary:    "a summary",
		Path:       "/notes/work",
		Visibility: domain.VisibilityPublic,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	pbEntry := domainEntryToProto(entry, []string{"tag-1", "tag-2"})
	assert.Equal(t, "entry-1", pbEntry.Id)
	assert.Equal(t, "user-1", pbEntry.UserId)
	assert.Equal(t, "Test Entry", pbEntry.Title)
	assert.Equal(t, "some content", pbEntry.Content)
	assert.Equal(t, "a summary", pbEntry.Summary)
	assert.Equal(t, "/notes/work", pbEntry.Path)
	assert.Equal(t, pb.Visibility_VISIBILITY_PUBLIC, pbEntry.Visibility)
	assert.Equal(t, []string{"tag-1", "tag-2"}, pbEntry.Tags)
	assert.Equal(t, now.Unix(), pbEntry.CreatedAt.AsTime().Unix())
}

func TestDomainEntryToProto_NilTags(t *testing.T) {
	entry := &domain.Entry{ID: "e1", Visibility: domain.VisibilityPrivate}
	pbEntry := domainEntryToProto(entry, nil)
	assert.Nil(t, pbEntry.Tags)
	assert.Equal(t, pb.Visibility_VISIBILITY_PRIVATE, pbEntry.Visibility)
}

// ---------------------------------------------------------------------------
// domainTagToProto
// ---------------------------------------------------------------------------

func TestDomainTagToProto(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	tag := &domain.Tag{
		ID:        "tag-1",
		Name:      "Golang",
		Slug:      "golang",
		CreatedAt: now,
	}

	pbTag := domainTagToProto(tag)
	assert.Equal(t, "tag-1", pbTag.Id)
	assert.Equal(t, "Golang", pbTag.Name)
	assert.Equal(t, "golang", pbTag.Slug)
	assert.Equal(t, now.Unix(), pbTag.CreatedAt.AsTime().Unix())
}

// ---------------------------------------------------------------------------
// protoToVisibility / visibilityToProto
// ---------------------------------------------------------------------------

func TestProtoToVisibility(t *testing.T) {
	tests := []struct {
		name     string
		input    pb.Visibility
		expected domain.Visibility
	}{
		{"public", pb.Visibility_VISIBILITY_PUBLIC, domain.VisibilityPublic},
		{"private", pb.Visibility_VISIBILITY_PRIVATE, domain.VisibilityPrivate},
		{"unspecified defaults to private", pb.Visibility_VISIBILITY_UNSPECIFIED, domain.VisibilityPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, protoToVisibility(tt.input))
		})
	}
}

func TestVisibilityToProto(t *testing.T) {
	tests := []struct {
		name     string
		input    domain.Visibility
		expected pb.Visibility
	}{
		{"public", domain.VisibilityPublic, pb.Visibility_VISIBILITY_PUBLIC},
		{"private", domain.VisibilityPrivate, pb.Visibility_VISIBILITY_PRIVATE},
		{"unknown defaults to unspecified", domain.Visibility(99), pb.Visibility_VISIBILITY_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, visibilityToProto(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// userIDFromContext
// ---------------------------------------------------------------------------

func TestUserIDFromContext_Present(t *testing.T) {
	md := metadata.New(map[string]string{"x-user-id": "user-42"})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	assert.Equal(t, "user-42", userIDFromContext(ctx))
}

func TestUserIDFromContext_Missing(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, userIDFromContext(ctx))
}

func TestUserIDFromContext_EmptyMetadata(t *testing.T) {
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	assert.Empty(t, userIDFromContext(ctx))
}

// ---------------------------------------------------------------------------
// LoggingInterceptor
// ---------------------------------------------------------------------------

func TestLoggingInterceptor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := LoggingInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	resp, err := interceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestLoggingInterceptor_WithError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := LoggingInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	_, err := interceptor(context.Background(), nil, info, handler)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

// ---------------------------------------------------------------------------
// RecoveryInterceptor
// ---------------------------------------------------------------------------

func TestRecoveryInterceptor_NoPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := RecoveryInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestRecoveryInterceptor_Panic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	interceptor := RecoveryInterceptor(logger)

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("unexpected error")
	}

	resp, err := interceptor(context.Background(), nil, info, handler)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
}
