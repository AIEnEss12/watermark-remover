package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/username/watermark-remover/internal/imgutil"
)

type ImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

// HealthCheck returns the status of the service.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RemoveWatermarkURL handles image processing from a URL.
func RemoveWatermarkURL(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Download image
	resp, err := http.Get(req.ImageURL)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to fetch image from URL"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Failed to fetch image: source returned non-200 status"})
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image data"})
		return
	}

	processAndRespond(c, data)
}

// RemoveWatermarkUpload handles image processing from a file upload.
func RemoveWatermarkUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	processAndRespond(c, data)
}

var (
	// Limit concurrent image processing tasks to prevent CPU/RAM exhaustion
	procSemaphore = make(chan struct{}, 4)
)

func processAndRespond(c *gin.Context, data []byte) {
	// Acquire semaphore
	procSemaphore <- struct{}{}
	defer func() { <-procSemaphore }()

	// Decode
	img, err := imgutil.DecodeImage(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image content"})
		return
	}
	defer img.Close()

	// Detect
	bboxes := imgutil.DetectWatermark(img)

	// Remove
	result, err := imgutil.RemoveWatermark(img, bboxes, "logo.png")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process image"})
		return
	}
	defer result.Close()

	// Encode
	format := c.Query("format")
	var resBytes []byte
	var contentType string

	if format == "webp" {
		resBytes, err = imgutil.EncodeToWebP(result, 80)
		contentType = "image/webp"
	} else {
		resBytes, err = imgutil.EncodeToJPEG(result, 90)
		contentType = "image/jpeg"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode result image"})
		return
	}

	c.Data(http.StatusOK, contentType, resBytes)
}
