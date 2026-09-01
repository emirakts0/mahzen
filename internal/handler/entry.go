package handler

import (
	"cmp"
	"log/slog"
	"net/http"
	"strings"
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

func newEntryHandler(svc *service.EntryService) *entryHandler {
	return &entryHandler{svc: svc}
}

// entryRequest is the JSON body for entry create/update.
type entryRequest struct {
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
		CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  e.UpdatedAt.Format(time.RFC3339),
	}
}

// respondWithEntry renders an entry with its tags as JSON.
func (h *entryHandler) respondWithEntry(c *gin.Context, entry *domain.Entry) {
	tags, err := h.svc.GetEntryTags(c.Request.Context(), entry.ID)
	if err != nil {
		slog.Warn("failed to fetch tags for entry", "entry_id", entry.ID, "error", err)
	}
	respondData(c, http.StatusOK, gin.H{"entry": domainEntryToResponse(entry, tags)})
}

func (h *entryHandler) createEntry(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	var req entryRequest
	if !bindJSON(c, &req) {
		return
	}

	entry, err := h.svc.CreateEntry(c.Request.Context(), userID, req.Title, req.Content, req.Path, req.FileType, domain.ParseVisibility(req.Visibility), req.TagIDs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "creating entry: "+err.Error())
		return
	}

	h.respondWithEntry(c, entry)
}

func (h *entryHandler) getEntry(c *gin.Context) {
	entry, err := h.svc.GetEntry(c.Request.Context(), c.Param("entry_id"))
	if err != nil {
		respondError(c, http.StatusNotFound, "entry not found: "+err.Error())
		return
	}

	h.respondWithEntry(c, entry)
}

func (h *entryHandler) updateEntry(c *gin.Context) {
	var req entryRequest
	if !bindJSON(c, &req) {
		return
	}

	entry, err := h.svc.UpdateEntry(c.Request.Context(), c.Param("entry_id"), req.Title, req.Content, req.Path, req.FileType, domain.ParseVisibility(req.Visibility), req.TagIDs)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "updating entry: "+err.Error())
		return
	}

	h.respondWithEntry(c, entry)
}

func (h *entryHandler) deleteEntry(c *gin.Context) {
	if err := h.svc.DeleteEntry(c.Request.Context(), c.Param("entry_id")); err != nil {
		respondError(c, http.StatusInternalServerError, "deleting entry: "+err.Error())
		return
	}

	respondData(c, http.StatusOK, gin.H{})
}

func (h *entryHandler) listEntries(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}

	limit, offset := parsePagination(c, 20)
	from, to := parseDateFilter(c.Query("from_date"), c.Query("to_date"))

	filter := &domain.ListEntriesFilter{
		Visibility: c.Query("visibility"),
		FromDate:   from,
		ToDate:     to,
	}
	if tags := c.Query("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	listPath := cmp.Or(c.Query("path"), "/")

	entries, folderInfos, total, err := h.svc.ListChildren(c.Request.Context(), userID, listPath, c.Query("own") == "true", filter, limit, offset)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "listing entries: "+err.Error())
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

	respondData(c, http.StatusOK, gin.H{
		"entries": items,
		"folders": folders,
		"total":   total,
	})
}
