package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// platformInfo maps platform names to file extension and content type.
var platformInfo = map[string]struct {
	ext         string
	contentType string
}{
	"windows": {ext: ".exe", contentType: "application/octet-stream"},
	"linux":   {ext: ".AppImage", contentType: "application/octet-stream"},
	"macos":   {ext: ".dmg", contentType: "application/octet-stream"},
}

// downloadHandler handles desktop app download requests.
type downloadHandler struct{}

// newDownloadHandler creates a new download handler.
func newDownloadHandler() *downloadHandler {
	return &downloadHandler{}
}

func (h *downloadHandler) download(c *gin.Context) {
	platform := c.Param("platform")

	info, ok := platformInfo[platform]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unsupported platform: %s (valid: windows, linux, macos)", platform),
		})
		return
	}

	if platform == "macos" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "macOS build coming soon",
		})
		return
	}

	filename := "mahzen-" + platform + info.ext
	content := fmt.Sprintf("Mahzen %s placeholder binary — real build coming soon.\n", platform)

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, info.contentType, []byte(content))
}
