package lyrics

import (
	"fmt"
	"net/http"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Fetcher retrieves lyrics from online sources
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new lyrics fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{},
	}
}

// FetchLyrics retrieves LRC lyrics for a song
func (f *Fetcher) FetchLyrics(song *models.SongInfo) (string, error) {
	// TODO: Step 5 - Implement LRCLIB API integration
	// LRCLIB API docs: https://lrclib.net/docs

	return "", fmt.Errorf("not implemented")
}
