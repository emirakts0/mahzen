package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// distFS embeds the compiled frontend assets.
// Populated by: cp -r web/dist/* cmd/mahzen/dist/
// The Makefile "dist" target handles this automatically.
//
//go:embed all:dist
var distFS embed.FS

// registerSPARoutes registers the embedded SPA handler on the Gin engine.
// API routes (/v1/) are NOT handled here — they go to the API handlers.
// Any request for a file that exists is served directly; everything else
// falls back to index.html so React Router can handle client-side routing.
func registerSPARoutes(r *gin.Engine) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic("failed to create sub filesystem for embedded frontend: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	r.NoRoute(func(c *gin.Context) {
		// Skip API routes — they should 404 normally.
		if strings.HasPrefix(c.Request.URL.Path, "/v1/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// Try to open the requested path in the embedded FS.
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if f, err := sub.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// File not found — serve index.html for SPA client-side routing.
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
