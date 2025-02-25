package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/dev3mike/video-stream-api/internal/service"
)

// VideoHandler handles HTTP requests for video operations
type VideoHandler struct {
	videoService *service.VideoService
}

// NewVideoHandler creates a new VideoHandler instance
func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
	}
}

// Upload handles video upload requests
func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Unable to read video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	videoID, err := h.videoService.ProcessVideo(header.Filename, header.Size, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"videoID": videoID})
}

// GetProgress handles progress check requests
func (h *VideoHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	progress := h.videoService.GetProgress(videoID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"progress": progress})
}

// StreamFile serves a full video file with Range support
func (h *VideoHandler) StreamFile(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	filePath := filepath.Join("videos", videoID+".mp4") // Assuming MP4 format

	http.ServeFile(w, r, filePath)
}

// StreamHLS serves HLS segments and playlists
func (h *VideoHandler) StreamHLS(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	filePath := filepath.Join("output", videoID)
	fileServer := http.StripPrefix(
		fmt.Sprintf("/video/%s/stream", videoID),
		http.FileServer(http.Dir(filePath)))
	fileServer.ServeHTTP(w, r)
}
