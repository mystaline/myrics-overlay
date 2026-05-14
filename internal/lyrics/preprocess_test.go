package lyrics

import (
	"testing"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

func TestCleanField_stripsBrackets(t *testing.T) {
	cases := []struct{ input, want string }{
		{"Song [Thai Sub]", "Song"},
		{"Song [Official MV]", "Song"},
		{"Song [Lyric Video]", "Song"},
		{"Song [Full]", "Song"},
	}
	for _, c := range cases {
		got := cleanField(c.input)
		if got != c.want {
			t.Errorf("cleanField(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestCleanField_stripsParenNoise(t *testing.T) {
	cases := []struct{ input, want string }{
		{"Song (Official MV)", "Song"},
		{"Song (Official Video)", "Song"},
		{"Song (Lyric Video)", "Song"},
		{"Song (MV)", "Song"},
		{"Song (PV)", "Song"},
		{"Song (Full Version)", "Song"},
		{"Song (Audio)", "Song"},
	}
	for _, c := range cases {
		got := cleanField(c.input)
		if got != c.want {
			t.Errorf("cleanField(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestCleanField_preservesFeat(t *testing.T) {
	input := "Song (feat. Artist)"
	got := cleanField(input)
	if got != input {
		t.Errorf("cleanField(%q) = %q, want unchanged", input, got)
	}
}

func TestCleanField_collapseSpaces(t *testing.T) {
	got := cleanField("Song   [MV]   Title")
	if got != "Song Title" {
		t.Errorf("cleanField: got %q, want %q", got, "Song Title")
	}
}

func TestCleanSong(t *testing.T) {
	song := &models.SongInfo{
		Artist: "Bell4210",
		Title:  "sana - Kimiiro RenAi Moyou [Thai Sub]",
	}
	cleaned := cleanSong(song)
	if cleaned.Title != "sana - Kimiiro RenAi Moyou" {
		t.Errorf("Title: got %q, want %q", cleaned.Title, "sana - Kimiiro RenAi Moyou")
	}
	if cleaned.Artist != "Bell4210" {
		t.Errorf("Artist: got %q, want %q", cleaned.Artist, "Bell4210")
	}
}

func TestSplitTitleArtist_withDash(t *testing.T) {
	song := &models.SongInfo{
		Artist: "Bell4210",
		Title:  "sana - Kimiiro RenAi Moyou",
	}
	result := splitTitleArtist(song)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Artist != "sana" {
		t.Errorf("Artist: got %q, want %q", result.Artist, "sana")
	}
	if result.Title != "Kimiiro RenAi Moyou" {
		t.Errorf("Title: got %q, want %q", result.Title, "Kimiiro RenAi Moyou")
	}
}

func TestSplitTitleArtist_noDash_returnsNil(t *testing.T) {
	song := &models.SongInfo{Artist: "Artist", Title: "Song Title"}
	if splitTitleArtist(song) != nil {
		t.Error("expected nil for title without ' - '")
	}
}

func TestSplitTitleArtist_stripsNoiseFromTitle(t *testing.T) {
	song := &models.SongInfo{
		Artist: "Channel",
		Title:  "Artist - Song [Official MV]",
	}
	result := splitTitleArtist(song)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Title != "Song" {
		t.Errorf("Title: got %q, want %q", result.Title, "Song")
	}
}
