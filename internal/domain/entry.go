package domain

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Visibility represents the access level of an entry.
type Visibility int

const (
	VisibilityPublic  Visibility = iota // Visible to all authenticated users.
	VisibilityPrivate                   // Visible only to the owning user.
)

// String returns the string representation of the visibility level.
func (v Visibility) String() string {
	switch v {
	case VisibilityPublic:
		return "public"
	case VisibilityPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// ParseVisibility converts a string to a Visibility value.
func ParseVisibility(s string) Visibility {
	switch s {
	case "public":
		return VisibilityPublic
	case "private":
		return VisibilityPrivate
	default:
		return VisibilityPrivate
	}
}

// Entry represents a knowledge entry stored in the platform.
type Entry struct {
	ID         string
	UserID     string
	Title      string
	Content    string // Text content stored directly in the database.
	Summary    string // AI-generated summary.
	Path       string // Materialized path for hierarchical organization (e.g. "/notes/work").
	Visibility Visibility
	FileType   string    // File extension provided by the client (e.g. "md", "txt"). Empty for plain text entries.
	FileSize   int64     // Size of the content in bytes.
	Embedding  []float32 // OpenAI embedding vector (stored as JSON in DB).
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NormalizePath cleans and validates a path string for use as an entry path.
// An empty string is normalized to "/" (root). Consecutive and trailing
// slashes are collapsed; "." and ".." segments are rejected.
func NormalizePath(p string) (string, error) {
	parts := strings.Split(strings.TrimSpace(p), "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		switch part {
		case ".", "..":
			return "", fmt.Errorf("path segment %q is not allowed", part)
		}
		if utf8.RuneCountInString(part) > 255 {
			return "", fmt.Errorf("path segment %q exceeds 255 characters", part)
		}
		for _, r := range part {
			if !isPathRune(r) {
				return "", fmt.Errorf("path contains invalid character %q", r)
			}
		}
		segments = append(segments, part)
	}

	if len(segments) == 0 {
		return "/", nil
	}

	p = "/" + strings.Join(segments, "/")
	if len(p) > 4096 {
		return "", fmt.Errorf("path exceeds maximum length of 4096 bytes")
	}

	return p, nil
}

// isPathRune reports whether r is allowed in a path segment.
// Allowed: letters, digits, hyphen, underscore, dot, tilde.
func isPathRune(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	switch r {
	case '-', '_', '.', '~':
		return true
	}
	return false
}

// EmbedText returns the text suitable for generating an embedding.
// It uses the content, falling back to title + summary.
func (e *Entry) EmbedText() string {
	if e.Content != "" {
		return e.Content
	}
	if e.Summary != "" {
		return e.Title + " " + e.Summary
	}
	return e.Title
}

// ListEntriesFilter contains optional filters for listing entries.
type ListEntriesFilter struct {
	Visibility string
	Tags       []string
	FromDate   time.Time
	ToDate     time.Time
}

// EntryRepository defines persistence operations for entries.
type EntryRepository interface {
	Create(ctx context.Context, entry *Entry) error
	GetByID(ctx context.Context, id string) (*Entry, error)
	Update(ctx context.Context, entry *Entry) error
	Delete(ctx context.Context, id string) error
	ListDistinctPaths(ctx context.Context, userID string) ([]string, error)
	ListInPath(ctx context.Context, userID, path string, own bool, filter *ListEntriesFilter, limit, offset int) ([]*Entry, int, error)
	ListPathCountsUnderPrefix(ctx context.Context, userID, prefix string, own bool, filter *ListEntriesFilter) ([]PathCount, error)
	ListAll(ctx context.Context) ([]*Entry, error)
	UpdateEmbedding(ctx context.Context, entryID string, embedding []float32) error
}

// PathCount holds a path and the number of entries at that path.
type PathCount struct {
	Path  string
	Count int
}
