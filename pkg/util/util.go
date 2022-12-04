package util

import (
	"fmt"
	"strings"
)

func ValidatePath(p string) error {
	if strings.Contains(p, "..") || strings.Contains(p, "~") {
		return fmt.Errorf("invalid path: %s must not contain '..' or '~'", p)
	}
	return nil
}

func FormatFileSize(numBytes int64) string {
	var unit string
	var conv int64
	if conv = 1024 * 1024 * 2014; numBytes > conv {
		unit = "GB"
	} else if conv = 1024 * 1024; numBytes > conv {
		unit = "MB"
	} else if conv = 1024; numBytes > conv {
		unit = "KB"
	} else {
		unit = "B"
		conv = 1
	}
	return fmt.Sprintf("%.2f%s", float64(numBytes)/float64(conv), unit)
}
