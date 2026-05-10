package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// SongInfo represents metadata about a detected song.
type SongInfo struct {
	Title    string
	Artist   string
	Album    string
	Duration int // in seconds
}

// LyricLine represents a single line of lyrics with timestamp.
type LyricLine struct {
	Timestamp time.Duration // offset from song start
	Text      string
}

// ACRCloudResponse represents the JSON response structure from ACRCloud.
type ACRCloudResponse struct {
	Status struct {
		Msg  string `json:"msg"`
		Code int    `json:"code"`
	} `json:"status"`
	Metadata struct {
		Music []struct {
			Title   string `json:"title"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Album struct {
				Name string `json:"name"`
			} `json:"album"`
			DurationMs int `json:"duration_ms"`
		} `json:"music"`
	} `json:"metadata"`
}

// ParseACRCloudResponse parses the JSON string from ACRCloud into SongInfo.
// Returns (nil, nil) when no music is detected (code 1001).
// Returns (nil, err) for API errors such as bad credentials or quota exceeded.
func ParseACRCloudResponse(jsonStr string) (*SongInfo, error) {
	var resp ACRCloudResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	switch resp.Status.Code {
	case 0:
		// Success — fall through to parse music metadata
	case 1001:
		// No music detected — expected when nothing is playing
		return nil, nil
	default:
		// API error: bad credentials (2000), quota exceeded (3003), invalid audio (2001), etc.
		return nil, fmt.Errorf("ACRCloud error %d: %s", resp.Status.Code, resp.Status.Msg)
	}

	if len(resp.Metadata.Music) == 0 {
		return nil, nil
	}

	track := resp.Metadata.Music[0]

	artist := "Unknown Artist"
	if len(track.Artists) > 0 {
		artist = track.Artists[0].Name
	}

	return &SongInfo{
		Title:    track.Title,
		Artist:   artist,
		Album:    track.Album.Name,
		Duration: track.DurationMs / 1000,
	}, nil
}
