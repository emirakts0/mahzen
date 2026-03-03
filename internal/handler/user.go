package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
)

// userHandler implements the user HTTP handlers.
type userHandler struct {
	userRepo domain.UserRepository
}

// newUserHandler creates a new userHandler.
func newUserHandler(userRepo domain.UserRepository) *userHandler {
	return &userHandler{userRepo: userRepo}
}

// userResponse is the JSON representation of a user.
type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func (h *userHandler) getCurrentUser(c *gin.Context) {
	userID := userIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": userResponse{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	})
}
