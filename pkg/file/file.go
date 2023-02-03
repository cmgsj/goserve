package file

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type Entry struct {
	Path     string
	Name     string
	Size     string
	IsDir    bool
	Children []*Entry
}

func (e *Entry) FindMatch(s string) (*Entry, error) {
	if e == nil {
		return nil, fmt.Errorf("nil entry")
	}
	parts := strings.Split(s, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		found := false
		for _, c := range e.Children {
			if c.Name == part {
				e = c
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("path not found: %s", s)
		}
	}
	return e, nil
}

func GetFSRoot(fileName string) (*Entry, error) {
	fstat, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	root := &Entry{
		Path: fileName,
		Name: fstat.Name(),
	}
	if fstat.IsDir() {
		root.Size = " - "
		root.IsDir = true
		entries, err := os.ReadDir(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		var files, dirs []*Entry
		for _, entry := range entries {
			finfo, err := entry.Info()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if finfo.IsDir() {
				f, err := GetFSRoot(path.Join(fileName, entry.Name()))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}
				dirs = append(dirs, f)
			} else {
				files = append(files, &Entry{
					Path:  path.Join(fileName, entry.Name()),
					Name:  entry.Name(),
					Size:  FormatSize(finfo.Size()),
					IsDir: false,
				})
			}
		}
		root.Children = append(dirs, files...)
	} else {
		root.Size = FormatSize(fstat.Size())
		root.IsDir = false
	}
	return root, nil
}

func FormatSize(fsize int64) string {
	var unit string
	var conv int64
	if conv = 1024 * 1024 * 2014; fsize > conv {
		unit = "GB"
	} else if conv = 1024 * 1024; fsize > conv {
		unit = "MB"
	} else if conv = 1024; fsize > conv {
		unit = "KB"
	} else {
		unit = "B"
		conv = 1
	}
	return fmt.Sprintf("%.2f%s", float64(fsize)/float64(conv), unit)
}
