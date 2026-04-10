package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/username/watermark-remover/internal/api"
)

func main() {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Routes
	r.GET("/health", api.HealthCheck)
	r.POST("/remove", api.RemoveWatermarkURL)
	r.POST("/remove/upload", api.RemoveWatermarkUpload)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting ENCAR Watermark Remover (Go) on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
