package lyrics

import (
	"regexp"
	"strings"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

var (
	reBrackets   = regexp.MustCompile(`\s*\[[^\]]*\]\s*`)
	reParenNoise = regexp.MustCompile(`(?i)\s*\((official[^)]*|lyric[^)]*|music\s*video|mv|pv|full[^)]*|audio[^)]*|ver(?:\.|sion)?[^)]*)\)\s*`)
)

// cleanSong strips YT Music noise (bracket tags, official video markers) from title and artist.
func cleanSong(song *models.SongInfo) *models.SongInfo {
	return &models.SongInfo{
		Artist: cleanField(song.Artist),
		Title:  cleanField(song.Title),
	}
}

func cleanField(s string) string {
	s = reBrackets.ReplaceAllString(s, " ")
	s = reParenNoise.ReplaceAllString(s, " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// splitTitleArtist handles the YT Music pattern where the channel is the SMTC artist
// and the real "Artist - Title" is embedded in the title field.
// Returns nil if title doesn't contain " - ".
func splitTitleArtist(song *models.SongInfo) *models.SongInfo {
	parts := strings.SplitN(song.Title, " - ", 2)
	if len(parts) < 2 {
		return nil
	}
	return &models.SongInfo{
		Artist: strings.TrimSpace(parts[0]),
		Title:  cleanField(strings.TrimSpace(parts[1])),
	}
}
