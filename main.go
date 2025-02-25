package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/dev3mike/video-stream-api/api/handlers"
	"github.com/dev3mike/video-stream-api/config"
	"github.com/dev3mike/video-stream-api/internal/models"
	"github.com/dev3mike/video-stream-api/internal/service"
	"github.com/dev3mike/video-stream-api/pkg/ffmpeg"
)

func main() {
	// Initialize configuration
	cfg := config.NewDefaultConfig()

	// Create required directories
	os.MkdirAll("videos", os.ModePerm)
	os.MkdirAll("output", os.ModePerm)

	// Validate FFmpeg installation
	validator := ffmpeg.NewValidator()
	if err := validator.Validate(); err != nil {
		cfg.Logger.Fatal("FFmpeg not found: ", err)
	}

	// Initialize components
	progress := models.NewProgress()
	videoService := service.NewVideoService(cfg, progress)
	videoHandler := handlers.NewVideoHandler(videoService)

	// Setup router
	r := chi.NewRouter()

	// Register routes
	r.Post("/upload", videoHandler.Upload)
	r.Get("/video/{id}/progress", videoHandler.GetProgress)
	r.Get("/video/{id}/streamfile", videoHandler.StreamFile)
	r.Get("/video/{id}/stream/*", videoHandler.StreamHLS)

	// Start the server
	cfg.Logger.Info("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		cfg.Logger.Fatal(err)
	}
}
