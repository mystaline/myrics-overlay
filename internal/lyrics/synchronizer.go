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
	// 1. Calculate elapsed time since startTime
	// 2. Find the line with the largest timestamp <= elapsed time
	// 3. Return that line (or nil if no match)

	elapsed := time.Since(s.startTime)

	var current *models.LyricLine
	for i := range s.lines {
		if s.lines[i].Timestamp <= elapsed {
			current = &s.lines[i]
		} else {
			break
		}
	}

	return current
}

// GetNextLine returns the upcoming lyric line
func (s *Synchronizer) GetNextLine() *models.LyricLine {
	// TODO: Step 7 - Get the next line after current
	elapsed := time.Since(s.startTime)

	for i := range s.lines {
		if s.lines[i].Timestamp > elapsed {
			return &s.lines[i]
		}
	}

	return nil
}

// Reset resets the synchronizer with a new start time
func (s *Synchronizer) Reset(startTime time.Time) {
	s.startTime = startTime
}
