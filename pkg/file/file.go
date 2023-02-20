package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("file not found")
)

type (
	Tree struct {
		Path     string
		Name     string
		Size     int64
		IsDir    bool
		IsBroken bool
		Children []*Tree
	}
	TreeStats struct {
		NumFiles  int
		TotalSize int64
		TimeDelta time.Duration
	}
)

func (t *Tree) FindMatch(filePath string) (*Tree, error) {
	if filePath == "/" {
		return t, nil
	}
	var (
		match = t
		parts = strings.Split(strings.Trim(filePath, "/"), "/")
	)
	for _, part := range parts {
		found := false
		for _, child := range match.Children {
			if child.Name == part {
				match = child
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%w: %q", ErrNotFound, filePath)
		}
	}
	return match, nil
}

func GetFileTree(filePath string, skipDotFiles bool, errWriter io.Writer) (*Tree, *TreeStats, error) {
	start := time.Now()
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, nil, err
	}
	var (
		root = &Tree{
			Path:  absPath,
			Name:  info.Name(),
			Size:  info.Size(),
			IsDir: info.IsDir(),
		}
		stats = &TreeStats{}
		queue = []*Tree{root}
	)
	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]
		if f.IsDir {
			entries, err := os.ReadDir(f.Path)
			if err != nil {
				f.IsBroken = true
				fmt.Fprintln(errWriter, err)
				continue
			}
			f.Children = make([]*Tree, 0, len(entries))
			for _, entry := range entries {
				if skipDotFiles && strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					fmt.Fprintln(errWriter, err)
					continue
				}
				child := &Tree{
					Path:  filepath.Join(f.Path, info.Name()),
					Name:  info.Name(),
					Size:  info.Size(),
					IsDir: info.IsDir(),
				}
				f.Children = append(f.Children, child)
				queue = append(queue, child)
			}
		} else {
			stats.TotalSize += f.Size
		}
		stats.NumFiles++
	}
	stats.TimeDelta = time.Since(start)
	return root, stats, nil
}
