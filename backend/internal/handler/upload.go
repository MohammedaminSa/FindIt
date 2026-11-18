package handler

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// UploadHandler manages HTTP multipart uploads and media lifecycle endpoints.
type UploadHandler struct {
	uploadService service.UploadService
}

// NewUploadHandler constructs a new UploadHandler.
func NewUploadHandler(uploadService service.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// RegisterRoutes mounts media upload and management routes under the provided router group.
func (h *UploadHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware ...gin.HandlerFunc) {
	uploads := rg.Group("/uploads")
	if len(authMiddleware) > 0 && authMiddleware[0] != nil {
		uploads.Use(authMiddleware[0])
	}
	{
		uploads.POST("", h.Upload)
		uploads.POST("/multiple", h.UploadMultiple)
		uploads.DELETE("/:filename", h.Delete)
	}

	// Mount convenient single-upload alias POST /api/v1/upload
	uploadAlias := rg.Group("/upload")
	if len(authMiddleware) > 0 && authMiddleware[0] != nil {
		uploadAlias.Use(authMiddleware[0])
	}
	{
		uploadAlias.POST("", h.Upload)
	}
}

// Upload handles single or multiple image uploads via multipart/form-data.
// Supported form keys: "images" (multiple/single), "image" (single), "file" (single).
// POST /api/v1/uploads or POST /api/v1/upload
func (h *UploadHandler) Upload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		// Attempt single file lookup under "image" or "file"
		file, errSingle := c.FormFile("image")
		if errSingle != nil {
			file, errSingle = c.FormFile("file")
		}
		if errSingle != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No image file provided in request",
			})
			return
		}

		url, uploadErr := h.uploadService.UploadImage(file)
		if uploadErr != nil {
			h.handleError(c, uploadErr)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"url":  url,
			"urls": []string{url},
		})
		return
	}

	// Look for files under "images", "image", or "file"
	var files []*multipart.FileHeader
	if f, ok := form.File["images"]; ok && len(f) > 0 {
		files = f
	} else if f, ok := form.File["image"]; ok && len(f) > 0 {
		files = f
	} else if f, ok := form.File["file"]; ok && len(f) > 0 {
		files = f
	}

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No image file provided in request",
		})
		return
	}

	if len(files) == 1 {
		url, uploadErr := h.uploadService.UploadImage(files[0])
		if uploadErr != nil {
			h.handleError(c, uploadErr)
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"url":  url,
			"urls": []string{url},
		})
		return
	}

	// Batch upload
	urls, uploadErr := h.uploadService.UploadImages(files)
	if uploadErr != nil {
		h.handleError(c, uploadErr)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"url":  urls[0],
		"urls": urls,
	})
}

// UploadMultiple handles explicit batch uploads.
// POST /api/v1/uploads/multiple
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid multipart form data",
		})
		return
	}

	var files []*multipart.FileHeader
	if f, ok := form.File["images"]; ok && len(f) > 0 {
		files = f
	} else if f, ok := form.File["image"]; ok && len(f) > 0 {
		files = f
	}

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No image files provided in request",
		})
		return
	}

	urls, uploadErr := h.uploadService.UploadImages(files)
	if uploadErr != nil {
		h.handleError(c, uploadErr)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"url":  urls[0],
		"urls": urls,
	})
}

// Delete handles deletion of an uploaded file.
// DELETE /api/v1/uploads/:filename
func (h *UploadHandler) Delete(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Filename is required",
		})
		return
	}

	if err := h.uploadService.DeleteImage(filename); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}

// handleError maps domain service errors to appropriate HTTP status codes and JSON messages.
func (h *UploadHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmptyFile):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Uploaded file is empty",
		})
	case errors.Is(err, service.ErrFileTooLarge):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size exceeds the 5MB limit",
		})
	case errors.Is(err, service.ErrInvalidFileType):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported file format. Only JPEG, PNG, and WebP images are allowed",
		})
	case errors.Is(err, service.ErrTooManyFiles):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 5 images allowed per upload",
		})
	case errors.Is(err, service.ErrInvalidFileName):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file name",
		})
	case errors.Is(err, service.ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Image not found",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process image upload",
		})
	}
}
