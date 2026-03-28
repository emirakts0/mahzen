package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// tagHandler implements the tag HTTP handlers.
type tagHandler struct {
	svc *service.TagService
}

// newTagHandler creates a new tagHandler.
func newTagHandler(svc *service.TagService) *tagHandler {
	return &tagHandler{svc: svc}
}

// createTagRequest is the JSON body for POST /v1/tags.
type createTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// attachTagRequest is the JSON body for POST /v1/entries/:entry_id/tags.
type attachTagRequest struct {
	TagID string `json:"tag_id" binding:"required"`
}

// tagResponse is the JSON representation of a tag.
type tagResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
}

func domainTagToResponse(t *domain.Tag) tagResponse {
	return tagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}

func (h *tagHandler) createTag(c *gin.Context) {
	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.svc.CreateTag(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "creating tag: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": domainTagToResponse(tag),
	})
}

func (h *tagHandler) getTag(c *gin.Context) {
	id := c.Param("id")

	tag, err := h.svc.GetTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tag": domainTagToResponse(tag),
	})
}

func (h *tagHandler) listTags(c *gin.Context) {
	limit, offset := parsePagination(c, 20)

	tags, total, err := h.svc.ListTags(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listing tags: " + err.Error()})
		return
	}

	items := make([]tagResponse, len(tags))
	for i, t := range tags {
		items[i] = domainTagToResponse(t)
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":  items,
		"total": total,
	})
}

func (h *tagHandler) deleteTag(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteTag(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "deleting tag: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *tagHandler) attachTag(c *gin.Context) {
	entryID := c.Param("entry_id")

	var req attachTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AttachTag(c.Request.Context(), entryID, req.TagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "attaching tag: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *tagHandler) detachTag(c *gin.Context) {
	entryID := c.Param("entry_id")
	tagID := c.Param("tag_id")

	if err := h.svc.DetachTag(c.Request.Context(), entryID, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "detaching tag: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
