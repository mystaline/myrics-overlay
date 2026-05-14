package lyrics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mystaline/myrics-overlay/internal/config"
	"github.com/mystaline/myrics-overlay/pkg/models"
)

// Fetcher retrieves lyrics from LRCLIB with NetEase as fallback.
type Fetcher struct {
	client    *http.Client
	lrclibURL string
	netease   *neteaseFetcher
}

// NewFetcher creates a new lyrics fetcher using URLs from config.
func NewFetcher(cfg *config.Config) *Fetcher {
	return &Fetcher{
		client:    &http.Client{Timeout: 10 * time.Second},
		lrclibURL: cfg.Lyrics.LRCLibURL,
		netease:   newNeteaseFetcher(cfg.Lyrics.NeteaseSearchURL, cfg.Lyrics.NeteaseLyricsURL),
	}
}

// FetchLyrics retrieves LRC lyrics for a song.
// Tries LRCLIB first, falls back to NetEase if not found.
// Respects ctx cancellation — returns immediately if the context is cancelled.
func (f *Fetcher) FetchLyrics(ctx context.Context, song *models.SongInfo) (string, error) {
	log.Printf("[lyrics] fetching: %s - %s", song.Artist, song.Title)

	cleaned := cleanSong(song)

	lyrics, err := f.fetchFromLRCLIB(ctx, cleaned)
	if err == nil {
		log.Printf("[lyrics] LRCLIB: found")
		return lyrics, nil
	}
	log.Printf("[lyrics] LRCLIB: %v", err)

	lyrics, neteaseErr := f.netease.fetchLyrics(ctx, cleaned)
	if neteaseErr == nil {
		log.Printf("[lyrics] NetEase: found")
		return lyrics, nil
	}
	log.Printf("[lyrics] NetEase: %v", neteaseErr)

	// YT Music often embeds "Real Artist - Title" inside the title field,
	// with the channel name as artist. Retry with the split.
	if split := splitTitleArtist(cleaned); split != nil {
		log.Printf("[lyrics] retrying with split title — artist: %q title: %q", split.Artist, split.Title)

		lyrics, err = f.fetchFromLRCLIB(ctx, split)
		if err == nil {
			log.Printf("[lyrics] LRCLIB (split): found")
			return lyrics, nil
		}

		lyrics, neteaseErr = f.netease.fetchLyrics(ctx, split)
		if neteaseErr == nil {
			log.Printf("[lyrics] NetEase (split): found")
			return lyrics, nil
		}
		log.Printf("[lyrics] NetEase (split): %v", neteaseErr)
	}

	return "", fmt.Errorf("lrclib: %v; netease: %v", err, neteaseErr)
}

func (f *Fetcher) fetchFromLRCLIB(ctx context.Context, song *models.SongInfo) (string, error) {
	params := url.Values{}
	params.Add("q", song.Artist+" "+song.Title)
	apiURL := f.lrclibURL + "?" + params.Encode()

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
		TrackName    string `json:"trackName"`
		ArtistName   string `json:"artistName"`
		SyncedLyrics string `json:"syncedLyrics"`
		PlainLyrics  string `json:"plainLyrics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no lyrics found")
	}

	titleNorm := strings.ToLower(strings.TrimSpace(song.Title))
	artistNorm := strings.ToLower(strings.TrimSpace(song.Artist))

	pick := func(synced bool) (string, bool) {
		// pass 1: exact title + artist
		for _, r := range results {
			if strings.EqualFold(strings.TrimSpace(r.TrackName), song.Title) &&
				strings.Contains(strings.ToLower(r.ArtistName), artistNorm) {
				if synced && r.SyncedLyrics != "" {
					log.Printf("[lrclib] exact match: %q by %q", r.TrackName, r.ArtistName)
					return r.SyncedLyrics, true
				}
				if !synced && r.PlainLyrics != "" {
					log.Printf("[lrclib] exact match (plain): %q by %q", r.TrackName, r.ArtistName)
					return r.PlainLyrics, true
				}
			}
		}
		// pass 2: title contains (but result title must not be longer by more than 10 chars)
		for _, r := range results {
			rTitle := strings.ToLower(strings.TrimSpace(r.TrackName))
			if strings.Contains(rTitle, titleNorm) &&
				len(rTitle)-len(titleNorm) <= 10 &&
				strings.Contains(strings.ToLower(r.ArtistName), artistNorm) {
				if synced && r.SyncedLyrics != "" {
					log.Printf("[lrclib] fuzzy match: %q by %q", r.TrackName, r.ArtistName)
					return r.SyncedLyrics, true
				}
				if !synced && r.PlainLyrics != "" {
					log.Printf("[lrclib] fuzzy match (plain): %q by %q", r.TrackName, r.ArtistName)
					return r.PlainLyrics, true
				}
			}
		}
		return "", false
	}

	if v, ok := pick(true); ok {
		return v, nil
	}
	if v, ok := pick(false); ok {
		return v, nil
	}

	log.Printf("[lrclib] no match in %d results for %q by %q", len(results), song.Title, song.Artist)
	return "", fmt.Errorf("no matching lyrics found")
}
