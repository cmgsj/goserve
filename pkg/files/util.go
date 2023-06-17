package files

import (
	"fmt"
	"time"

	"github.com/cmgsj/goserve/pkg/templates"
)

const (
	BytesCopiedHeader = "X-Bytes-Copied"
)

type FileSlice []templates.File

func (s FileSlice) Len() int {
	return len(s)
}

func (s FileSlice) Less(i, j int) bool {
	if s[i].IsDir == s[j].IsDir {
		return s[i].Name < s[j].Name
	}
	return s[i].IsDir
}

func (s FileSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func formatFileSize(size int64) string {
	var (
		unit   string
		factor int64
	)
	if factor = 1024 * 1024 * 1024; size >= factor {
		unit = "GB"
	} else if factor = 1024 * 1024; size >= factor {
		unit = "MB"
	} else if factor = 1024; size >= factor {
		unit = "KB"
	} else {
		unit = "B"
		factor = 1
	}
	return fmt.Sprintf("%0.2f%s", float64(size)/float64(factor), unit)
}

func formatDuration(d time.Duration) string {
	var (
		unit   string
		factor int64
		n      = d.Nanoseconds()
	)
	if factor = 60 * 1000 * 1000 * 1000; n >= factor {
		unit = "min"
	} else if factor = 1000 * 1000 * 1000; n >= factor {
		unit = "s"
	} else if factor = 1000 * 1000; n >= factor {
		unit = "ms"
	} else if factor = 1000; n >= factor {
		unit = "µs"
	} else {
		unit = "ns"
		factor = 1
	}
	return fmt.Sprintf("%.2f%s", float64(n)/float64(factor), unit)
}
