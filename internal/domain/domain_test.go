package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Visibility
// ---------------------------------------------------------------------------

func TestVisibility_String(t *testing.T) {
	tests := []struct {
		name     string
		v        Visibility
		expected string
	}{
		{"public", VisibilityPublic, "public"},
		{"private", VisibilityPrivate, "private"},
		{"unknown", Visibility(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.v.String())
		})
	}
}

func TestParseVisibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Visibility
	}{
		{"public", "public", VisibilityPublic},
		{"private", "private", VisibilityPrivate},
		{"unknown defaults to private", "unknown", VisibilityPrivate},
		{"empty defaults to private", "", VisibilityPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseVisibility(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// Visibility enum values
// ---------------------------------------------------------------------------

func TestVisibilityEnumValues(t *testing.T) {
	// Ensures the iota values are stable.
	assert.Equal(t, Visibility(0), VisibilityPublic)
	assert.Equal(t, Visibility(1), VisibilityPrivate)
}

// ---------------------------------------------------------------------------
// NormalizePath
// ---------------------------------------------------------------------------

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		// Valid paths.
		{"empty string becomes root", "", "/", false},
		{"slash becomes root", "/", "/", false},
		{"simple path", "/notes", "/notes", false},
		{"nested path", "/notes/work/meeting", "/notes/work/meeting", false},
		{"path without leading slash", "notes/work", "/notes/work", false},
		{"trailing slash removed", "/notes/work/", "/notes/work", false},
		{"multiple trailing slashes", "/notes///", "/notes", false},
		{"consecutive slashes collapsed", "/notes//work///meeting", "/notes/work/meeting", false},
		{"whitespace trimmed", "  /notes  ", "/notes", false},
		{"allowed special chars", "/my-notes_2026/v1.0~draft", "/my-notes_2026/v1.0~draft", false},
		{"single segment", "/a", "/a", false},
		{"only slashes becomes root", "///", "/", false},

		// Invalid paths.
		{"dot segment", "/notes/./work", "", true},
		{"double-dot segment", "/notes/../work", "", true},
		{"invalid character space", "/notes/my work", "", true},
		{"invalid character at-sign", "/notes/t@g", "", true},
		{"invalid character hash", "/notes/t#g", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizePath(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePath_SegmentTooLong(t *testing.T) {
	longSegment := strings.Repeat("a", 256)
	_, err := NormalizePath("/" + longSegment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds 255 characters")
}

func TestNormalizePath_PathTooLong(t *testing.T) {
	// Build a path that exceeds 4096 bytes using many valid segments.
	var segments []string
	for i := 0; i < 500; i++ {
		segments = append(segments, strings.Repeat("x", 10))
	}
	longPath := "/" + strings.Join(segments, "/")
	assert.True(t, len(longPath) > 4096)

	_, err := NormalizePath(longPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum length")
}
