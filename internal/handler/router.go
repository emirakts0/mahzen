package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/emirakts0/mahzen/internal/domain"
	"github.com/emirakts0/mahzen/internal/service"
)

// RouterDeps holds all dependencies required to set up API routes.
type RouterDeps struct {
	Logger   *slog.Logger
	TokenGen domain.TokenGenerator

	EntrySvc  *service.EntryService
	TagSvc    *service.TagService
	SearchSvc *service.SearchService
	AuthSvc   *service.AuthService
	UserRepo  domain.UserRepository
}

// SetupRouter creates a Gin engine with all middleware and routes registered.
func SetupRouter(deps RouterDeps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// Global middleware.
	r.Use(CORSMiddleware())
	r.Use(RecoveryMiddleware(deps.Logger))
	r.Use(LoggingMiddleware(deps.Logger))
	r.Use(AuthMiddleware(deps.Logger, deps.TokenGen))

	// Create handlers.
	auth := newAuthHandler(deps.AuthSvc)
	entries := newEntryHandler(deps.EntrySvc)
	tags := newTagHandler(deps.TagSvc)
	search := newSearchHandler(deps.SearchSvc)
	users := newUserHandler(deps.UserRepo)

	// API routes under /v1.
	v1 := r.Group("/v1")
	{
		// Auth routes (public).
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", auth.register)
			authGroup.POST("/login", auth.login)
			authGroup.POST("/refresh", auth.refreshToken)
			authGroup.POST("/logout", auth.logout)
		}

		// Entry routes.
		entryGroup := v1.Group("/entries")
		{
			entryGroup.GET("", entries.listEntries)
			entryGroup.GET("/:entry_id", entries.getEntry)
			entryGroup.POST("", entries.createEntry)
			entryGroup.PUT("/:entry_id", entries.updateEntry)
			entryGroup.DELETE("/:entry_id", entries.deleteEntry)

			// Entry-tag relationship routes.
			entryGroup.POST("/:entry_id/tags", tags.attachTag)
			entryGroup.DELETE("/:entry_id/tags/:tag_id", tags.detachTag)
		}

		// Tag routes.
		tagGroup := v1.Group("/tags")
		{
			tagGroup.POST("", tags.createTag)
			tagGroup.GET("", tags.listTags)
			tagGroup.GET("/:id", tags.getTag)
			tagGroup.DELETE("/:id", tags.deleteTag)
		}

		// Search routes (public).
		searchGroup := v1.Group("/search")
		{
			searchGroup.GET("/keyword", search.keywordSearch)
			searchGroup.GET("/semantic", search.semanticSearch)
		}

		// User routes.
		userGroup := v1.Group("/users")
		{
			userGroup.GET("/me", users.getCurrentUser)
		}
	}

	return r
}
