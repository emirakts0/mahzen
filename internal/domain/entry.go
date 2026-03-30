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
// An empty string is normalized to "/" (root). The path must start with "/",
// must not end with "/" (except root), and may only contain URL-safe characters.
func NormalizePath(p string) (string, error) {
	p = strings.TrimSpace(p)

	// Empty or just "/" → root.
	if p == "" || p == "/" {
		return "/", nil
	}

	// Must start with "/".
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	// Remove trailing slash.
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/", nil
	}

	// Collapse consecutive slashes in a single pass.
	parts := strings.Split(p, "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) == 0 {
		return "/", nil
	}
	p = "/" + strings.Join(filtered, "/")

	// Validate each segment.
	segments := strings.Split(p[1:], "/") // skip leading "/"
	for _, seg := range segments {
		if seg == "" {
			return "", fmt.Errorf("path contains empty segment")
		}
		if seg == "." || seg == ".." {
			return "", fmt.Errorf("path segment %q is not allowed", seg)
		}
		if utf8.RuneCountInString(seg) > 255 {
			return "", fmt.Errorf("path segment %q exceeds 255 characters", seg)
		}
		for _, r := range seg {
			if !isPathRune(r) {
				return "", fmt.Errorf("path contains invalid character %q", r)
			}
		}
	}

	// Overall path length check.
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
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Entry, int, error)
	ListAccessible(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*Entry, int, error)
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
