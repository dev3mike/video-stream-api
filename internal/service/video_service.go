package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/dev3mike/video-stream-api/config"
	"github.com/dev3mike/video-stream-api/internal/models"
	"github.com/dev3mike/video-stream-api/pkg/ffmpeg"
)

// VideoService handles video-related operations
type VideoService struct {
	config   *config.Config
	progress *models.Progress
	logger   *logrus.Logger
}

// NewVideoService creates a new VideoService instance
func NewVideoService(cfg *config.Config, progress *models.Progress) *VideoService {
	return &VideoService{
		config:   cfg,
		progress: progress,
		logger:   cfg.Logger,
	}
}

// ProcessVideo handles video upload and transcoding
func (s *VideoService) ProcessVideo(filename string, fileSize int64, file multipart.File) (string, error) {
	// Validate file size
	if fileSize > s.config.MaxFileSize {
		return "", fmt.Errorf("file too large")
	}

	// Validate file type
	fileExt := strings.ToLower(filepath.Ext(filename))
	if fileExt == "" {
		return "", fmt.Errorf("file has no extension")
	}
	fileExt = fileExt[1:] // Remove the dot

	validType := false
	for _, allowedType := range s.config.AllowedFileTypes {
		if fileExt == allowedType {
			validType = true
			break
		}
	}
	if !validType {
		return "", fmt.Errorf("invalid file type: %s", fileExt)
	}

	// Generate UUID for video ID
	videoID := uuid.New().String()
	filePath := filepath.Join("videos", videoID+"."+fileExt)

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy the file
	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Get video duration
	duration, err := ffmpeg.GetVideoDuration(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get video duration: %w", err)
	}

	// Initialize progress to 0
	s.progress.Update(videoID, 0)
	s.logger.WithFields(logrus.Fields{
		"videoID":  videoID,
		"duration": duration,
		"fileSize": fileSize,
		"fileName": filename,
	}).Info("Starting video transcoding")

	// Start transcoding in background
	go s.transcodeVideo(videoID, filePath, duration)

	return videoID, nil
}

// GetProgress returns the transcoding progress for a video
func (s *VideoService) GetProgress(videoID string) int {
	progress, exists := s.progress.Get(videoID)
	if !exists {
		return -1 // Return -1 to indicate video not found
	}
	return progress
}

func (s *VideoService) transcodeVideo(videoID, inputPath string, durationMs int64) {
	outputDir := filepath.Join("output", videoID)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		s.logger.WithError(err).Error("Failed to create output directory")
		s.progress.Update(videoID, -1)
		return
	}

	var lastProgress int = 0
	transcoder := ffmpeg.NewTranscoder()
	err := transcoder.Transcode(ffmpeg.TranscodeOptions{
		VideoID:   videoID,
		InputPath: inputPath,
		OutputDir: outputDir,
		Config:    s.config.FFmpegConfig,
		Logger:    s.logger,
		OnProgress: func(videoID string, percent int) {
			// Only update if progress has increased
			if percent > lastProgress {
				s.logger.WithFields(logrus.Fields{
					"videoID":      videoID,
					"progress":     percent,
					"lastProgress": lastProgress,
				}).Debug("Transcoding progress update")
				s.progress.Update(videoID, percent)
				lastProgress = percent
			}
		},
		DurationMs: durationMs,
	})

	if err != nil {
		s.logger.WithError(err).Error("Transcoding failed")
		s.progress.Update(videoID, -1)
		return
	}

	// Only set to 100 if we haven't encountered any errors
	s.logger.WithField("videoID", videoID).Info("Transcoding completed successfully")
	s.progress.Update(videoID, 100)
}
