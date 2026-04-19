package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/username/watermark-remover/internal/imgutil"
	"gocv.io/x/gocv"
)

type ImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Job queue for worker pool
	jobQueue = make(chan Job, 100)
)

type Job struct {
	Data      []byte
	Format    string // "remove" | "swap"
	OutputFmt string // "jpeg" | "webp"
	CacheKey  string // Empty for uploads
	ReplyCh   chan Result
}

type Result struct {
	Data        []byte
	ContentType string
	Err         error
}

func init() {
	// Start workers based on CPU count
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
}

func worker() {
	for job := range jobQueue {
		res := Result{}
		
		// Decode
		img, err := imgutil.DecodeImage(job.Data)
		if err != nil {
			res.Err = fmt.Errorf("invalid image content")
			job.ReplyCh <- res
			continue
		}

		var result gocv.Mat
		if job.Format == "remove" {
			// Detect
			bboxes := imgutil.DetectWatermark(img)
			// Remove
			result, err = imgutil.RemoveWatermark(img, bboxes, "logo.png")
		} else {
			// Swap (just pass through to encoder which handles it)
			result = img.Clone()
		}

		if err != nil {
			img.Close()
			res.Err = fmt.Errorf("failed to process image")
			job.ReplyCh <- res
			continue
		}

		// Encode
		var resBytes []byte
		var contentType string
		if job.OutputFmt == "webp" {
			resBytes, err = imgutil.EncodeToWebP(result, 80)
			contentType = "image/webp"
		} else {
			resBytes, err = imgutil.EncodeToJPEG(result, 90)
			contentType = "image/jpeg"
		}

		img.Close()
		result.Close()

		if err != nil {
			res.Err = fmt.Errorf("failed to encode result")
		} else {
			res.Data = resBytes
			res.ContentType = contentType
		}
		
		job.ReplyCh <- res
	}
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

	data, contentType, err := downloadImage(req.ImageURL)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if res, ok := GetCachedImage(req.ImageURL + "remove" + c.Query("format")); ok {
		c.Data(http.StatusOK, res.ContentType, res.Data)
		return
	}

	submitJob(c, data, "remove", contentType, req.ImageURL+"remove"+c.Query("format"))
}

// SwapColorsURL handles R/B channel swapping from a URL.
func SwapColorsURL(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	data, contentType, err := downloadImage(req.ImageURL)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if res, ok := GetCachedImage(req.ImageURL + "swap" + c.Query("format")); ok {
		c.Data(http.StatusOK, res.ContentType, res.Data)
		return
	}

	submitJob(c, data, "swap", contentType, req.ImageURL+"swap"+c.Query("format"))
}

// RemoveWatermarkUpload handles image processing from a file upload.
func RemoveWatermarkUpload(c *gin.Context) {
	data, err := readUploadedFile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submitJob(c, data, "remove", "", "")
}

// SwapColorsUpload handles R/B channel swapping from a file upload.
func SwapColorsUpload(c *gin.Context) {
	data, err := readUploadedFile(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submitJob(c, data, "swap", "", "")
}

func submitJob(c *gin.Context, data []byte, format string, originalContentType string, cacheKey string) {
	// Check if already processed
	if imgutil.HasProcessingTag(data) {
		ct := originalContentType
		if ct == "" {
			ct = http.DetectContentType(data)
		}
		c.Data(http.StatusOK, ct, data)
		return
	}

	replyCh := make(chan Result, 1)
	job := Job{
		Data:      data,
		Format:    format,
		OutputFmt: c.Query("format"),
		CacheKey:  cacheKey,
		ReplyCh:   replyCh,
	}

	select {
	case jobQueue <- job:
		// Wait for result with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		select {
		case res := <-replyCh:
			if res.Err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": res.Err.Error()})
			} else {
				if job.CacheKey != "" {
					AddCachedImage(job.CacheKey, res.Data, res.ContentType)
				}
				c.Data(http.StatusOK, res.ContentType, res.Data)
			}
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Processing timed out"})
		}
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server is busy, try again later"})
	}
}

// Helpers

func downloadImage(url string) ([]byte, string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image from URL")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: source returned %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data")
	}

	return data, contentType, nil
}

func readUploadedFile(c *gin.Context) ([]byte, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("no file uploaded")
	}

	if file.Size > 20<<20 { // 20 MB limit
		return nil, fmt.Errorf("file too large (max 20MB)")
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file")
	}
	defer f.Close()

	return io.ReadAll(f)
}
