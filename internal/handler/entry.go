package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// folderResponse represents folder information in the API response.
type folderResponse struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

// entryHandler implements the entry HTTP handlers.
type entryHandler struct {
	svc *service.EntryService
}

// newEntryHandler creates a new entryHandler.
func newEntryHandler(svc *service.EntryService) *entryHandler {
	return &entryHandler{svc: svc}
}

// createEntryRequest is the JSON body for POST /v1/entries.
type createEntryRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Path       string   `json:"path"`
	Visibility string   `json:"visibility"`
	TagIDs     []string `json:"tag_ids"`
	FileType   string   `json:"file_type"`
}

// updateEntryRequest is the JSON body for PUT /v1/entries/:id.
type updateEntryRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Path       string   `json:"path"`
	Visibility string   `json:"visibility"`
	TagIDs     []string `json:"tag_ids"`
	FileType   string   `json:"file_type"`
}

// entryResponse is the JSON representation of an entry.
type entryResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary,omitempty"`
	Path       string   `json:"path"`
	Visibility string   `json:"visibility"`
	Tags       []string `json:"tags,omitempty"`
	FileType   string   `json:"file_type,omitempty"`
	FileSize   int64    `json:"file_size,omitempty"`
	S3Key      string   `json:"s3_key,omitempty"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

func domainEntryToResponse(e *domain.Entry, tags []string) *entryResponse {
	return &entryResponse{
		ID:         e.ID,
		UserID:     e.UserID,
		Title:      e.Title,
		Content:    e.Content,
		Summary:    e.Summary,
		Path:       e.Path,
		Visibility: e.Visibility.String(),
		Tags:       tags,
		FileType:   e.FileType,
		FileSize:   e.FileSize,
		S3Key:      e.S3Key,
		CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  e.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *entryHandler) createEntry(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req createEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vis := domain.ParseVisibility(req.Visibility)

	entry, err := h.svc.CreateEntry(c.Request.Context(), userID, req.Title, req.Content, req.Path, req.FileType, vis, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "creating entry: " + err.Error()})
		return
	}

	tags, err := h.svc.GetEntryTags(c.Request.Context(), entry.ID)
	if err != nil {
		slog.Warn("failed to fetch tags for created entry", "entry_id", entry.ID, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, tags),
	})
}

func (h *entryHandler) getEntry(c *gin.Context) {
	id := c.Param("entry_id")

	entry, err := h.svc.GetEntry(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found: " + err.Error()})
		return
	}

	tags, err := h.svc.GetEntryTags(c.Request.Context(), id)
	if err != nil {
		slog.Warn("failed to fetch tags for entry", "entry_id", id, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, tags),
	})
}

func (h *entryHandler) updateEntry(c *gin.Context) {
	id := c.Param("entry_id")

	var req updateEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vis := domain.ParseVisibility(req.Visibility)

	entry, err := h.svc.UpdateEntry(c.Request.Context(), id, req.Title, req.Content, req.Path, req.FileType, vis, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "updating entry: " + err.Error()})
		return
	}

	tags, err := h.svc.GetEntryTags(c.Request.Context(), id)
	if err != nil {
		slog.Warn("failed to fetch tags for updated entry", "entry_id", id, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, tags),
	})
}

func (h *entryHandler) deleteEntry(c *gin.Context) {
	id := c.Param("entry_id")

	if err := h.svc.DeleteEntry(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "deleting entry: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *entryHandler) listEntries(c *gin.Context) {
	userID := userIDFromContext(c)
	path := c.Query("path")
	own := c.Query("own") == "true"

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v >= 0 {
			limit = v
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// When path is specified, use ListChildren which returns entries AND folders
	if path != "" {
		entries, folderInfos, total, err := h.svc.ListChildren(c.Request.Context(), userID, path, own, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "listing children: " + err.Error()})
			return
		}

		entryIDs := make([]string, len(entries))
		for i, e := range entries {
			entryIDs[i] = e.ID
		}
		tagsByEntry, err := h.svc.GetEntryTagsBatch(c.Request.Context(), entryIDs)
		if err != nil {
			slog.Warn("failed to batch fetch tags for entries", "error", err)
			tagsByEntry = map[string][]string{}
		}

		items := make([]*entryResponse, len(entries))
		for i, e := range entries {
			items[i] = domainEntryToResponse(e, tagsByEntry[e.ID])
		}

		folders := make([]folderResponse, len(folderInfos))
		for i, f := range folderInfos {
			folders[i] = folderResponse{Path: f.Path, Count: f.Count}
		}

		c.JSON(http.StatusOK, gin.H{
			"entries": items,
			"folders": folders,
			"total":   total,
		})
		return
	}

	// No path specified - list all entries and root-level folders with counts
	rootPath := path
	if rootPath == "" {
		rootPath = "/"
	}
	entries, folderInfos, total, err := h.svc.ListChildren(c.Request.Context(), userID, rootPath, own, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listing entries: " + err.Error()})
		return
	}

	entryIDs := make([]string, len(entries))
	for i, e := range entries {
		entryIDs[i] = e.ID
	}
	tagsByEntry, err := h.svc.GetEntryTagsBatch(c.Request.Context(), entryIDs)
	if err != nil {
		slog.Warn("failed to batch fetch tags for entries", "error", err)
		tagsByEntry = map[string][]string{}
	}

	items := make([]*entryResponse, len(entries))
	for i, e := range entries {
		items[i] = domainEntryToResponse(e, tagsByEntry[e.ID])
	}

	folders := make([]folderResponse, len(folderInfos))
	for i, f := range folderInfos {
		folders[i] = folderResponse{Path: f.Path, Count: f.Count}
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": items,
		"folders": folders,
		"total":   total,
	})
}

func (h *entryHandler) getDownloadURL(c *gin.Context) {
	id := c.Param("entry_id")

	url, err := h.svc.GetEntryDownloadURL(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
