package format

import (
	"fmt"
	"time"
)

func ThousandsSeparator(num int) string {
	s := fmt.Sprintf("%d", num)
	n := len(s)
	k := n + (n-1)/3
	out := make([]byte, k)
	for i := n - 1; i >= 0; i-- {
		if i != n-1 && (n-i-1)%3 == 0 {
			k--
			out[k] = ','
		}
		k--
		out[k] = s[i]
	}
	return string(out)
}

func FileSize(size int64) string {
	var unit string
	var factor int64
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

func TimeDuration(t time.Duration) string {
	var unit string
	var factor int64
	n := t.Nanoseconds()
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
