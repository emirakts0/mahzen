package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

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
}

// updateEntryRequest is the JSON body for PUT /v1/entries/:id.
type updateEntryRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Path       string   `json:"path"`
	Visibility string   `json:"visibility"`
	TagIDs     []string `json:"tag_ids"`
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

	entry, err := h.svc.CreateEntry(c.Request.Context(), userID, req.Title, req.Content, req.Path, vis, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "creating entry: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, nil),
	})
}

func (h *entryHandler) getEntry(c *gin.Context) {
	id := c.Param("entry_id")

	entry, err := h.svc.GetEntry(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, nil),
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

	entry, err := h.svc.UpdateEntry(c.Request.Context(), id, req.Title, req.Content, req.Path, vis, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "updating entry: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entry": domainEntryToResponse(entry, nil),
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

	limit := 20
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	entries, total, err := h.svc.ListEntries(c.Request.Context(), userID, path, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listing entries: " + err.Error()})
		return
	}

	items := make([]*entryResponse, len(entries))
	for i, e := range entries {
		items[i] = domainEntryToResponse(e, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": items,
		"total":   total,
	})
}
