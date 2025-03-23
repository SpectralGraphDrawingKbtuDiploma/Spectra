package mtxparser

import (
	"regexp"
	"strings"
)

// ParseDimensions extracts dimensions from MTX file content
func ParseDimensions(content string) string {
	// This is a simplified parser - you may need a more robust one
	// depending on the MTX format you're working with

	// Example: looking for lines like "%%MatrixMarket matrix coordinate real general"
	// and following dimensions line
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, "%%MatrixMarket") && i+1 < len(lines) {
			// Try to get dimensions from the next non-comment line
			for j := i + 1; j < len(lines); j++ {
				if !strings.HasPrefix(lines[j], "%") {
					// Found a non-comment line, try to parse dimensions
					re := regexp.MustCompile(`\s*(\d+)\s+(\d+)`)
					matches := re.FindStringSubmatch(lines[j])
					if len(matches) >= 3 {
						return matches[1] + "x" + matches[2]
					}
					break
				}
			}
		}
	}

	return "unknown"
}
