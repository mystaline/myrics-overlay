package lyrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// newIntegrationFetcher wires a Fetcher with controllable LRCLIB and NetEase servers.
func newIntegrationFetcher(lrclibURL, neteaseSearchURL, neteaseLyricsURL string) *Fetcher {
	return &Fetcher{
		client:    &http.Client{},
		lrclibURL: lrclibURL,
		netease:   newNeteaseFetcher(neteaseSearchURL, neteaseLyricsURL),
	}
}

// TestFetchLyrics_LRCLIBSucceeds verifies LRCLIB result is used when available.
func TestFetchLyrics_LRCLIBSucceeds(t *testing.T) {
	lrclib := lrclibServer([]lrclibResult{
		{TrackName: "Song", ArtistName: "Artist", SyncedLyrics: "[00:01.00]from lrclib"},
	})
	defer lrclib.Close()

	f := newIntegrationFetcher(lrclib.URL, "http://invalid", "http://invalid")
	got, err := f.FetchLyrics(context.Background(), &models.SongInfo{Artist: "Artist", Title: "Song"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[00:01.00]from lrclib" {
		t.Errorf("got %q, want lrclib result", got)
	}
}

// TestFetchLyrics_FallsBackToNetEase verifies NetEase is tried when LRCLIB has no match.
func TestFetchLyrics_FallsBackToNetEase(t *testing.T) {
	lrclib := lrclibServer([]lrclibResult{}) // no results

	neteaseSearch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"songs": []map[string]any{
					{"id": 12345, "name": "Song", "artists": []map[string]any{{"name": "Artist"}}},
				},
			},
			"code": 200,
		})
	}))
	neteaseLyrics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"lrc":     map[string]any{"lyric": "[00:01.00]from netease"},
			"nolyric": false,
			"code":    200,
		})
	}))
	defer lrclib.Close()
	defer neteaseSearch.Close()
	defer neteaseLyrics.Close()

	f := newIntegrationFetcher(lrclib.URL, neteaseSearch.URL, neteaseLyrics.URL)
	got, err := f.FetchLyrics(context.Background(), &models.SongInfo{Artist: "Artist", Title: "Song"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[00:01.00]from netease" {
		t.Errorf("got %q, want netease result", got)
	}
}

// TestFetchLyrics_SplitTitleFallback covers the YT Music pattern where the channel
// name is the SMTC artist and the real "Artist - Title" is embedded in the title.
// Regression for: Bell4210 / "sana - Kimiiro RenAi Moyou [Thai Sub]" not finding lyrics.
func TestFetchLyrics_SplitTitleFallback(t *testing.T) {
	// Both LRCLIB calls fail (channel query + split query for first attempt)
	callCount := 0
	lrclib := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode([]lrclibResult{}) // no results
	}))

	// NetEase fails on channel query, succeeds on split query
	neteaseSearchCallCount := 0
	neteaseSearch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		neteaseSearchCallCount++
		q := r.FormValue("s")
		if q == "sana Kimiiro RenAi Moyou" {
			json.NewEncoder(w).Encode(map[string]any{
				"result": map[string]any{
					"songs": []map[string]any{{"id": 99, "name": "Kimiiro RenAi Moyou"}},
				},
				"code": 200,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"songs": []any{}}, "code": 200})
		}
	}))
	neteaseLyrics := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"lrc":     map[string]any{"lyric": "[00:01.00]kimiiro lyrics"},
			"nolyric": false,
			"code":    200,
		})
	}))
	defer lrclib.Close()
	defer neteaseSearch.Close()
	defer neteaseLyrics.Close()

	f := newIntegrationFetcher(lrclib.URL, neteaseSearch.URL, neteaseLyrics.URL)
	song := &models.SongInfo{
		Artist: "Bell4210",
		Title:  "sana - Kimiiro RenAi Moyou [Thai Sub]",
	}
	got, err := f.FetchLyrics(context.Background(), song)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[00:01.00]kimiiro lyrics" {
		t.Errorf("got %q, want split-title result", got)
	}
}

// TestFetchLyrics_AllSourcesFail returns an error when both LRCLIB and NetEase have nothing.
func TestFetchLyrics_AllSourcesFail(t *testing.T) {
	lrclib := lrclibServer([]lrclibResult{})
	neteaseSearch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"songs": []any{}}, "code": 200})
	}))
	defer lrclib.Close()
	defer neteaseSearch.Close()

	f := newIntegrationFetcher(lrclib.URL, neteaseSearch.URL, "http://invalid")
	_, err := f.FetchLyrics(context.Background(), &models.SongInfo{Artist: "Unknown", Title: "Unknown"})
	if err == nil {
		t.Error("expected error when all sources fail")
	}
}
