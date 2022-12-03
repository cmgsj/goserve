package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	port = flag.Int("port", 1234, "port to listen on")
	root = flag.String("root", ".", "root path to serve")
	tmpl = template.Must(template.New("index").Parse(indexHTML))
	//go:embed index.html
	indexHTML string
	fsize     = int64(0)
)

func main() {
	flag.Parse()
	if !IsValidPath(*root) {
		fmt.Fprintln(os.Stderr, "root path cannot contain '..' or '~'")
		os.Exit(1)
	}
	*root = path.Clean(*root)
	fstat, err := os.Stat(*root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "path %s does not exist\n", *root)
		} else {
			fmt.Fprintf(os.Stderr, "error reading path %s: %v\n", *root, err)
		}
		os.Exit(1)
	}
	var fileType string
	if fstat.IsDir() {
		fileType = "dir"
		http.Handle("/", LoggerMiddleware(DirHandler))
	} else {
		fileType = "file"
		fsize = fstat.Size()
		http.Handle("/", LoggerMiddleware(FileHandler))
	}
	addr := fmt.Sprintf(":%d", *port)
	serverUrl := fmt.Sprintf("http://localhost%s", addr)
	fmt.Printf("serving %s [%s] at %s\n", fileType, *root, serverUrl)
	go OpenBrowser(serverUrl)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		tmpl.Execute(w, DocumentTemplate{
			Ok:       true,
			BackLink: "/",
			Header:   "/",
			Files: []FileTemplate{
				{
					Path:  *root,
					Name:  path.Base(*root),
					Size:  GetFileSize(fsize),
					IsDir: false,
				},
			},
		})
		return
	}
	SendFile(w, r, *root)
}

func DirHandler(w http.ResponseWriter, r *http.Request) {
	file := *root + r.URL.Path
	if !IsValidPath(file) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	fstat, err := os.Stat(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			tmpl.Execute(w, DocumentTemplate{
				Ok:       false,
				BackLink: "/",
				Header:   fmt.Sprintf("file not found: %s", file),
			})
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if !fstat.IsDir() {
		SendFile(w, r, file)
		return
	}
	entries, err := os.ReadDir(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var files []FileTemplate
	for _, entry := range entries {
		size := " - "
		if info, err := entry.Info(); err == nil && !entry.IsDir() {
			size = GetFileSize(info.Size())
		}
		files = append(files, FileTemplate{
			Path:  path.Join(r.URL.Path, entry.Name()),
			Name:  entry.Name(),
			Size:  size,
			IsDir: entry.IsDir(),
		})
	}
	tmpl.Execute(w, DocumentTemplate{
		Ok:       true,
		BackLink: path.Dir(strings.TrimRight(r.URL.Path, "/")),
		Header:   r.URL.Path,
		Files:    files,
	})
}

func SendFile(w http.ResponseWriter, r *http.Request, filePath string) {
	if !IsValidPath(filePath) {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	fileName := path.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	n, err := io.Copy(w, f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.Header.Set("bytes-copied", fmt.Sprintf("%d", n))
}

func LoggerMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := NewStatusRecorder(w)
		next(w, r)
		if n := r.Header.Get("bytes-copied"); n != "" {
			log.Printf("%s %s [%s] --> %s [%s bytes] %dms\n", r.Method, r.URL.Path, r.RemoteAddr, rec.Status, n, rec.Milis())
		} else {
			log.Printf("%s %s [%s] --> %s %dms\n", r.Method, r.URL.Path, r.RemoteAddr, rec.Status, rec.Milis())
		}
	})
}

func GetFileSize(numBytes int64) string {
	var conv float64
	var unit string
	if numBytes > 1024*1024*1024 {
		unit = "GB"
		conv = float64(numBytes) / 1024 / 1024 / 1024
	} else if numBytes > 1024*1024 {
		unit = "MB"
		conv = float64(numBytes) / 1024 / 1024
	} else if numBytes > 1024 {
		unit = "KB"
		conv = float64(numBytes) / 1024
	} else {
		unit = "B"
		conv = float64(numBytes)
	}
	return fmt.Sprintf("%.2f%s", conv, unit)
}

func IsValidPath(p string) bool {
	return !strings.Contains(p, "..") && !strings.Contains(p, "~")
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

type DocumentTemplate struct {
	Ok       bool
	BackLink string
	Header   string
	Files    []FileTemplate
}

type FileTemplate struct {
	Path  string
	Name  string
	Size  string
	IsDir bool
}

type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
	Status     string
	Start      time.Time
}

func (s *StatusRecorder) WriteHeader(code int) {
	s.StatusCode = code
	s.Status = http.StatusText(code)
	s.ResponseWriter.WriteHeader(code)
}

func (s *StatusRecorder) Milis() int64 {
	return time.Since(s.Start).Nanoseconds() / 1000 / 1000
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
		Status:         http.StatusText(http.StatusOK),
		Start:          time.Now(),
	}
}
