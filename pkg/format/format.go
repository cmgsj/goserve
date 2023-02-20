package format

import (
	"fmt"
	"strconv"
	"time"
)

func ThousandsSeparator(num int) string {
	var (
		s   = strconv.Itoa(num)
		n   = len(s)
		k   = n + (n-1)/3
		sep = make([]byte, k)
	)
	for i := n - 1; i >= 0; i-- {
		if i != n-1 && (n-i-1)%3 == 0 {
			k--
			sep[k] = ','
		}
		k--
		sep[k] = s[i]
	}
	return string(sep)
}

func FileSize(size int64) string {
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

func Duration(t time.Duration) string {
	var (
		unit   string
		factor int64
		n      = t.Nanoseconds()
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
