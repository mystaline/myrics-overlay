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

	"github.com/mystaline/myrics-overlay/pkg/models"
)

const (
	neteaseUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	neteaseReferer   = "https://music.163.com/"
)

type neteaseFetcher struct {
	client    *http.Client
	searchURL string
	lyricsURL string
}

func newNeteaseFetcher(searchURL, lyricsURL string) *neteaseFetcher {
	return &neteaseFetcher{
		client:    &http.Client{Timeout: 10 * time.Second},
		searchURL: searchURL,
		lyricsURL: lyricsURL,
	}
}

func (n *neteaseFetcher) fetchLyrics(ctx context.Context, song *models.SongInfo) (string, error) {
	songID, err := n.searchSong(ctx, song)
	if err != nil {
		return "", err
	}
	return n.fetchByID(ctx, songID)
}

func (n *neteaseFetcher) searchSong(ctx context.Context, song *models.SongInfo) (int64, error) {
	query := song.Artist + " " + song.Title
	log.Printf("[netease] searching: %q", query)

	form := url.Values{}
	form.Set("s", query)
	form.Set("type", "1")
	form.Set("limit", "5")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.searchURL, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", neteaseUserAgent)
	req.Header.Set("Referer", neteaseReferer)

	resp, err := n.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[netease] search status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("search status %d", resp.StatusCode)
	}

	var result struct {
		Result struct {
			Songs []struct {
				ID      int64  `json:"id"`
				Name    string `json:"name"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"songs"`
		} `json:"result"`
		Code int `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode search: %w", err)
	}
	if result.Code != 200 || len(result.Result.Songs) == 0 {
		log.Printf("[netease] no songs found (code=%d, count=%d)", result.Code, len(result.Result.Songs))
		return 0, fmt.Errorf("song not found on netease")
	}

	top := result.Result.Songs[0]
	artistName := ""
	if len(top.Artists) > 0 {
		artistName = top.Artists[0].Name
	}
	log.Printf("[netease] matched: %q by %q (ID: %d)", top.Name, artistName, top.ID)
	return top.ID, nil
}

func (n *neteaseFetcher) fetchByID(ctx context.Context, songID int64) (string, error) {
	apiURL := fmt.Sprintf("%s?id=%d&lv=-1&tv=-1", n.lyricsURL, songID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", neteaseUserAgent)
	req.Header.Set("Referer", neteaseReferer)

	resp, err := n.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("lyrics request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[netease] lyrics status: %d for ID %d", resp.StatusCode, songID)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lyrics status %d", resp.StatusCode)
	}

	var result struct {
		Lrc struct {
			Lyric string `json:"lyric"`
		} `json:"lrc"`
		NoLyric bool `json:"nolyric"`
		Code    int  `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode lyrics: %w", err)
	}
	if result.NoLyric || result.Lrc.Lyric == "" {
		return "", fmt.Errorf("no lyrics on netease")
	}

	return result.Lrc.Lyric, nil
}
