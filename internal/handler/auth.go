package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/service"
)

// authHandler implements the auth HTTP handlers.
type authHandler struct {
	authSvc *service.AuthService
}

// newAuthHandler creates a new authHandler.
func newAuthHandler(authSvc *service.AuthService) *authHandler {
	return &authHandler{authSvc: authSvc}
}

// registerRequest is the JSON body for POST /v1/auth/register.
type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password" binding:"required,min=8"`
}

// loginRequest is the JSON body for POST /v1/auth/login.
type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// refreshTokenRequest is the JSON body for POST /v1/auth/refresh.
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// logoutRequest is the JSON body for POST /v1/auth/logout.
type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// authResponse is the JSON response for auth endpoints.
type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *authHandler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authSvc.Register(c.Request.Context(), req.Email, req.DisplayName, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *authHandler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *authHandler) refreshToken(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token refresh failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *authHandler) logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
