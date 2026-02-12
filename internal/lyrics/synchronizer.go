package lyrics

import (
	"time"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Synchronizer tracks playback and returns current lyrics
type Synchronizer struct {
	lines     []models.LyricLine
	startTime time.Time
}

// NewSynchronizer creates a new lyrics synchronizer
func NewSynchronizer(lines []models.LyricLine, startTime time.Time) *Synchronizer {
	return &Synchronizer{
		lines:     lines,
		startTime: startTime,
	}
}

// GetCurrentLine returns the lyric line that should be displayed now
func (s *Synchronizer) GetCurrentLine() *models.LyricLine {
	// TODO: Step 7 - Implement synchronization logic
	return nil
}

// GetNextLine returns the upcoming lyric line
func (s *Synchronizer) GetNextLine() *models.LyricLine {
	// TODO: Step 7 - Get the next line after current
	return nil
}

// Reset resets the synchronizer with a new start time
func (s *Synchronizer) Reset(startTime time.Time) {
	s.startTime = startTime
}
