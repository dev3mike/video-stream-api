package models

import "sync"

// Progress represents the transcoding progress of a video
type Progress struct {
	store map[string]int
	mu    sync.RWMutex
}

// NewProgress creates a new Progress instance
func NewProgress() *Progress {
	return &Progress{
		store: make(map[string]int),
	}
}

// Update updates the progress for a video
func (p *Progress) Update(videoID string, percent int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.store[videoID] = percent
}

// Get retrieves the progress for a video
func (p *Progress) Get(videoID string) (int, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	progress, ok := p.store[videoID]
	return progress, ok
}
