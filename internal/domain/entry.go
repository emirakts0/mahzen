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
	Content    string     // Inline content when size is below S3 threshold.
	Summary    string     // AI-generated summary.
	S3Key      string     // Set when content is stored in object storage.
	Path       string     // Materialized path for hierarchical organization (e.g. "/notes/work").
	Visibility Visibility
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsStoredInS3 reports whether the entry's content resides in object storage.
func (e *Entry) IsStoredInS3() bool {
	return e.S3Key != ""
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

	// Collapse consecutive slashes.
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}

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

// EntryRepository defines persistence operations for entries.
type EntryRepository interface {
	Create(ctx context.Context, entry *Entry) error
	GetByID(ctx context.Context, id string) (*Entry, error)
	Update(ctx context.Context, entry *Entry) error
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Entry, int, error)
	ListAccessible(ctx context.Context, userID, pathPrefix string, limit, offset int) ([]*Entry, int, error)
}
