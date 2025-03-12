package modules

import (
	"fmt"
	"strings"
	"time"
)

// FormatDuration formats a duration as HH:MM:SS
func FormatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	hours := seconds / 3600
	seconds -= hours * 3600
	minutes := seconds / 60
	seconds -= minutes * 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// UpdateProgressInPlace updates the progress indicator on the same line with elapsed time
func UpdateProgressInPlace(current, total int, startTime time.Time) {
	percent := float64(current) / float64(total) * 100
	progressBar := MakeProgressBar(percent, 30)
	elapsed := time.Since(startTime)
	fmt.Printf("\r[%s] %3.1f%% Sending chunk: %d/%d | Elapsed: %s",
		progressBar, percent, current, total, FormatDuration(elapsed))
}

// MakeProgressBar creates a text-based progress bar
func MakeProgressBar(percent float64, width int) string {
	completed := int(percent / 100 * float64(width))
	remaining := width - completed

	bar := strings.Repeat("█", completed) + strings.Repeat("░", remaining)
	return bar
}
