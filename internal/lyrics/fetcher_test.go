package lyrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

type lrclibResult struct {
	TrackName    string `json:"trackName"`
	ArtistName   string `json:"artistName"`
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

func newTestFetcher(serverURL string) *Fetcher {
	return &Fetcher{
		client:    &http.Client{},
		lrclibURL: serverURL,
		netease:   newNeteaseFetcher("http://invalid", "http://invalid"),
	}
}

func lrclibServer(results []lrclibResult) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(results)
	}))
}

func TestFetchFromLRCLIB_exactMatch(t *testing.T) {
	srv := lrclibServer([]lrclibResult{
		{TrackName: "Story", ArtistName: "Kana Nishino", SyncedLyrics: "[00:01.00]line1"},
		{TrackName: "Bedtime Story", ArtistName: "Kana Nishino", SyncedLyrics: "[00:01.00]wrong"},
	})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	song := &models.SongInfo{Artist: "Kana Nishino", Title: "Story"}
	got, err := f.fetchFromLRCLIB(context.Background(), song)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "[00:01.00]line1" {
		t.Errorf("got wrong lyrics: %q", got)
	}
}

func TestFetchFromLRCLIB_rejectsBedtimeStory(t *testing.T) {
	// "Bedtime Story" should NOT match when querying "Story" — title too long (+8 chars but contains)
	// However exact match pass 1 should also reject it since titles differ
	srv := lrclibServer([]lrclibResult{
		{TrackName: "Bedtime Story", ArtistName: "Kana Nishino", SyncedLyrics: "[00:01.00]wrong"},
	})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	song := &models.SongInfo{Artist: "Kana Nishino", Title: "Story"}
	_, err := f.fetchFromLRCLIB(context.Background(), song)
	if err == nil {
		t.Error("expected no match for 'Bedtime Story' when querying 'Story'")
	}
}

func TestFetchFromLRCLIB_prefersSync(t *testing.T) {
	srv := lrclibServer([]lrclibResult{
		{TrackName: "Song", ArtistName: "Artist", PlainLyrics: "plain text", SyncedLyrics: "[00:01.00]synced"},
	})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	song := &models.SongInfo{Artist: "Artist", Title: "Song"}
	got, err := f.fetchFromLRCLIB(context.Background(), song)
	if err != nil {
		t.Fatal(err)
	}
	if got != "[00:01.00]synced" {
		t.Errorf("should prefer synced, got %q", got)
	}
}

func TestFetchFromLRCLIB_fallsBackToPlain(t *testing.T) {
	srv := lrclibServer([]lrclibResult{
		{TrackName: "Song", ArtistName: "Artist", PlainLyrics: "plain text"},
	})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	song := &models.SongInfo{Artist: "Artist", Title: "Song"}
	got, err := f.fetchFromLRCLIB(context.Background(), song)
	if err != nil {
		t.Fatal(err)
	}
	if got != "plain text" {
		t.Errorf("should fall back to plain, got %q", got)
	}
}

func TestFetchFromLRCLIB_emptyResults(t *testing.T) {
	srv := lrclibServer([]lrclibResult{})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	_, err := f.fetchFromLRCLIB(context.Background(), &models.SongInfo{Artist: "A", Title: "B"})
	if err == nil {
		t.Error("expected error for empty results")
	}
}

func TestFetchFromLRCLIB_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	_, err := f.fetchFromLRCLIB(context.Background(), &models.SongInfo{Artist: "A", Title: "B"})
	if err == nil {
		t.Error("expected error for 404")
	}
}

func TestFetchFromLRCLIB_fuzzyMatch(t *testing.T) {
	// "Song +" contains "Song" and is only 2 chars longer — within threshold
	srv := lrclibServer([]lrclibResult{
		{TrackName: "Song +", ArtistName: "Artist", SyncedLyrics: "[00:01.00]line"},
	})
	defer srv.Close()

	f := newTestFetcher(srv.URL)
	song := &models.SongInfo{Artist: "Artist", Title: "Song"}
	_, err := f.fetchFromLRCLIB(context.Background(), song)
	if err != nil {
		t.Errorf("expected fuzzy match for 'Song +' when querying 'Song': %v", err)
	}
}
