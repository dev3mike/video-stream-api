package config

import "github.com/sirupsen/logrus"

// Config holds the application configuration
type Config struct {
	MaxFileSize      int64    `json:"maxFileSize"`      // in bytes
	AllowedFileTypes []string `json:"allowedFileTypes"` // e.g., ["mp4", "mov", "avi"]
	FFmpegConfig     FFmpegConfig
	Logger           *logrus.Logger
}

type FFmpegConfig struct {
	Renditions      []Rendition `json:"renditions"`
	SegmentDuration int         `json:"segmentDuration"`
	PresetSpeed     string      `json:"presetSpeed"`
}

type Rendition struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Bitrate string `json:"bitrate"`
}

// NewDefaultConfig returns a new Config instance with default values
func NewDefaultConfig() *Config {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &Config{
		MaxFileSize:      500 * 1024 * 1024, // 500MB default
		AllowedFileTypes: []string{"mp4", "mov", "avi", "mkv"},
		FFmpegConfig: FFmpegConfig{
			Renditions: []Rendition{
				{Width: 640, Height: 360, Bitrate: "800k"},
				// {Width: 854, Height: 480, Bitrate: "1200k"},
				// {Width: 1280, Height: 720, Bitrate: "2000k"},
			},
			SegmentDuration: 4,
			PresetSpeed:     "fast",
		},
		Logger: logger,
	}
}
