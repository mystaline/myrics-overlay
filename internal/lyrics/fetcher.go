package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Fetcher retrieves lyrics from LRCLIB.
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new lyrics fetcher.
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// FetchLyrics retrieves LRC lyrics for a song from LRCLIB.
// Returns synced LRC content when available, plain text otherwise.
func (f *Fetcher) FetchLyrics(song *models.SongInfo) (string, error) {
	params := url.Values{}
	params.Add("q", song.Artist+" "+song.Title)
	apiURL := "https://lrclib.net/api/search?" + params.Encode()

	resp, err := f.client.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to call LRCLIB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("lyrics not found")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LRCLIB returned status: %d", resp.StatusCode)
	}

	var results []struct {
		SyncedLyrics string `json:"syncedLyrics"`
		PlainLyrics  string `json:"plainLyrics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no lyrics found")
	}

	result := results[0]
	if result.SyncedLyrics != "" {
		return result.SyncedLyrics, nil
	}
	if result.PlainLyrics != "" {
		return result.PlainLyrics, nil
	}

	return "", fmt.Errorf("no lyrics content found")
}
