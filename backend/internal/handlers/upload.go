package handlers

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"blog-platform-backend/internal/config"
	"blog-platform-backend/internal/storage"
)

// UploadHandler handles file upload and serving.
type UploadHandler struct {
	storage *storage.MinIOStorage
	cfg     *config.Config
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(s *storage.MinIOStorage, cfg *config.Config) *UploadHandler {
	return &UploadHandler{storage: s, cfg: cfg}
}

// allowedExtensions defines the set of permitted image file extensions.
var allowedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// maxUploadSize is the maximum allowed file size in bytes (5 MB).
const maxUploadSize int64 = 5 * 1024 * 1024

// Upload handles image file uploads.
func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large: maximum size is 5 MB"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType, ok := allowedExtensions[ext]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type: allowed types are jpg, png, gif, webp"})
		return
	}

	// Generate unique filename
	uuid := generateUUID()
	objectName := fmt.Sprintf("blog-images/%s%s", uuid, ext)

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Upload to MinIO
	if err := h.storage.Upload(c.Request.Context(), strings.NewReader(string(data)), objectName, contentType, int64(len(data))); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	// Strip the blog-images/ prefix for the proxy URL
	proxyFilename := strings.TrimPrefix(objectName, "blog-images/")

	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("/api/v1/files/%s", proxyFilename),
	})
}

// ServeFile proxies a file from MinIO to the client.
func (h *UploadHandler) ServeFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}

	objectName := "blog-images/" + filename

	obj, err := h.storage.Get(c.Request.Context(), objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve file"})
		return
	}
	defer obj.Close()

	// Stat to get content type and verify object exists
	stat, err := obj.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Data(http.StatusOK, contentType, data)
}

// generateUUID creates a random UUID v4 string.
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback should basically never fail, but handle it
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}
	// Set version 4 bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

