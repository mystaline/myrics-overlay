package lyrics

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mystaline/myrics-overlay/pkg/models"
)

// ParseLRC parses LRC format lyrics into structured lines sorted by timestamp.
func ParseLRC(lrcContent string) ([]models.LyricLine, error) {
	var lines []models.LyricLine
	scanner := bufio.NewScanner(strings.NewReader(lrcContent))

	// Matches [mm:ss.xx] or [mm:ss.xxx]
	timestampRegex := regexp.MustCompile(`\[(\d+):(\d+(\.\d+)?)\]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		matches := timestampRegex.FindAllStringSubmatchIndex(line, -1)
		if matches == nil {
			continue
		}

		lastMatchEnd := matches[len(matches)-1][1]
		content := strings.TrimSpace(line[lastMatchEnd:])

		for _, match := range matches {
			start, end := match[0], match[1]
			timestampStr := line[start+1 : end-1]

			duration, err := parseTimestamp(timestampStr)
			if err != nil {
				continue
			}

			lines = append(lines, models.LyricLine{
				Timestamp: duration,
				Text:      content,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid lyrics found")
	}

	// Sort ascending — LRC files are usually sorted, but not guaranteed.
	// The frontend sync loop relies on ascending order for its early-break optimisation.
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Timestamp < lines[j].Timestamp
	})

	return lines, nil
}

func parseTimestamp(ts string) (time.Duration, error) {
	parts := strings.Split(ts, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid timestamp format")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	totalSeconds := float64(minutes*60) + seconds
	return time.Duration(totalSeconds * float64(time.Second)), nil
}
