package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func IsValidPath(p string) bool {
	return !strings.Contains(p, "..") && !strings.Contains(p, "~")
}

func GetFileSize(numBytes int64) string {
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

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening browser: %v\n", err)
	}
}
