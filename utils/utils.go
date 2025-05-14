package utils

import (
	"fmt"
)

// FormatSize converts bytes to a human-readable string.
func FormatSize(sizeBytes int64) string {
	const (
		_  = iota
		kb = 1 << (10 * iota)
		mb
		gb
		tb
	)

	switch {
	case sizeBytes < kb:
		return fmt.Sprintf("%d B", sizeBytes)
	case sizeBytes < mb:
		return fmt.Sprintf("%.1f KB", float64(sizeBytes)/float64(kb))
	case sizeBytes < gb:
		return fmt.Sprintf("%.1f MB", float64(sizeBytes)/float64(mb))
	default: // Could add TB etc.
		return fmt.Sprintf("%.1f GB", float64(sizeBytes)/float64(gb))
	}
}
