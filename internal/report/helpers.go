package report

import (
	"fmt"
	"strings"
	"time"
)

func formatCarbonMg(carbonKg float64) string {
	carbonMg := carbonKg * 1_000_000
	if carbonMg < 0.01 {
		return fmt.Sprintf("%.4f", carbonMg)
	}
	if carbonMg < 1 {
		return fmt.Sprintf("%.3f", carbonMg)
	}
	if carbonMg < 100 {
		return fmt.Sprintf("%.2f", carbonMg)
	}
	return fmt.Sprintf("%.1f", carbonMg)
}

func formatBytes(value uint64) string {
	const unit = 1024
	if value < unit {
		return fmt.Sprintf("%d B", value)
	}

	div, exp := uint64(unit), 0
	for n := value / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	prefixes := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(prefixes) {
		exp = len(prefixes) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(value)/float64(div), prefixes[exp])
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}

	seconds := int64(duration.Seconds())
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func truncateMiddle(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	keep := max - 3
	left := keep / 2
	right := keep - left
	if left < 1 {
		left = 1
	}
	if right < 1 {
		right = 1
	}
	return value[:left] + "..." + value[len(value)-right:]
}
