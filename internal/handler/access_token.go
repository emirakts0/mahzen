package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/service"
)

// accessTokenHandler implements the access token HTTP handlers.
type accessTokenHandler struct {
	svc *service.AccessTokenService
}

// newAccessTokenHandler creates a new accessTokenHandler.
func newAccessTokenHandler(svc *service.AccessTokenService) *accessTokenHandler {
	return &accessTokenHandler{svc: svc}
}

type createTokenRequest struct {
	Name      string  `json:"name" binding:"required"`
	ExpiresIn *string `json:"expires_in,omitempty"` // e.g. "720h" for 30 days
}

type tokenResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

type createTokenResponse struct {
	Token    tokenResponse `json:"token"`
	RawToken string        `json:"raw_token"`
}

func (h *accessTokenHandler) createToken(c *gin.Context) {
	var req createTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := userIDFromContext(c)

	var expiresIn *time.Duration
	if req.ExpiresIn != nil {
		d, err := time.ParseDuration(*req.ExpiresIn)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_in duration"})
			return
		}
		expiresIn = &d
	}

	token, rawToken, err := h.svc.CreateToken(c.Request.Context(), userID, req.Name, expiresIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createTokenResponse{
		Token: tokenResponse{
			ID:        token.ID,
			Name:      token.Name,
			Prefix:    token.Prefix,
			Status:    token.Status,
			ExpiresAt: token.ExpiresAt.Format(time.RFC3339),
			CreatedAt: token.CreatedAt.Format(time.RFC3339),
		},
		RawToken: rawToken,
	})
}

func (h *accessTokenHandler) listTokens(c *gin.Context) {
	userID := userIDFromContext(c)

	tokens, err := h.svc.ListTokens(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tokens"})
		return
	}

	items := make([]tokenResponse, len(tokens))
	for i, t := range tokens {
		items[i] = tokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			Prefix:    t.Prefix,
			Status:    t.Status,
			ExpiresAt: t.ExpiresAt.Format(time.RFC3339),
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, gin.H{"tokens": items})
}

func (h *accessTokenHandler) revokeToken(c *gin.Context) {
	userID := userIDFromContext(c)
	tokenID := c.Param("id")

	if err := h.svc.RevokeToken(c.Request.Context(), userID, tokenID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
