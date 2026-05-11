package lyrics

import (
	"context"
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
// Respects ctx cancellation — returns immediately if the context is cancelled.
func (f *Fetcher) FetchLyrics(ctx context.Context, song *models.SongInfo) (string, error) {
	params := url.Values{}
	params.Add("q", song.Artist+" "+song.Title)
	apiURL := "https://lrclib.net/api/search?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	resp, err := f.client.Do(req)
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

	// Prefer the first result that has synced lyrics, then fall back to plain.
	for _, r := range results {
		if r.SyncedLyrics != "" {
			return r.SyncedLyrics, nil
		}
	}
	for _, r := range results {
		if r.PlainLyrics != "" {
			return r.PlainLyrics, nil
		}
	}

	return "", fmt.Errorf("no lyrics content found")
}
