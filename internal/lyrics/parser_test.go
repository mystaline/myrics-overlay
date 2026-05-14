package lyrics

import (
	"testing"
	"time"
)

func TestParseLRC_basic(t *testing.T) {
	input := "[00:12.34]Hello world\n[00:25.67]Second line\n"
	lines, err := ParseLRC(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d", len(lines))
	}
	if lines[0].Text != "Hello world" {
		t.Errorf("want %q, got %q", "Hello world", lines[0].Text)
	}
}

func TestParseLRC_sortedAscending(t *testing.T) {
	input := "[01:00.00]Late line\n[00:10.00]Early line\n"
	lines, err := ParseLRC(input)
	if err != nil {
		t.Fatal(err)
	}
	if lines[0].Text != "Early line" {
		t.Errorf("expected first line to be %q, got %q", "Early line", lines[0].Text)
	}
}

func TestParseLRC_multipleTimestampsPerLine(t *testing.T) {
	input := "[00:10.00][00:20.00]Chorus\n"
	lines, err := ParseLRC(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("want 2 lines for repeated timestamp, got %d", len(lines))
	}
	for _, l := range lines {
		if l.Text != "Chorus" {
			t.Errorf("want %q, got %q", "Chorus", l.Text)
		}
	}
}

func TestParseLRC_plainText_returnsError(t *testing.T) {
	input := "This is plain text\nNo timestamps here\n"
	_, err := ParseLRC(input)
	if err == nil {
		t.Error("expected error for plain text, got nil")
	}
}

func TestParseLRC_empty_returnsError(t *testing.T) {
	_, err := ParseLRC("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParseTimestamp(t *testing.T) {
	cases := []struct {
		input string
		want  time.Duration
	}{
		{"00:12.34", 12*time.Second + 340*time.Millisecond},
		{"01:05.00", 65 * time.Second},
		{"02:30.500", 150*time.Second + 500*time.Millisecond},
	}
	for _, c := range cases {
		got, err := parseTimestamp(c.input)
		if err != nil {
			t.Errorf("parseTimestamp(%q): unexpected error: %v", c.input, err)
			continue
		}
		// allow 1ms tolerance for float rounding
		diff := got - c.want
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Millisecond {
			t.Errorf("parseTimestamp(%q): want %v, got %v", c.input, c.want, got)
		}
	}
}

func TestParseTimestamp_invalid(t *testing.T) {
	cases := []string{"", "12", "ab:cd", "1:2:3"}
	for _, c := range cases {
		_, err := parseTimestamp(c)
		if err == nil {
			t.Errorf("parseTimestamp(%q): expected error, got nil", c)
		}
	}
}
