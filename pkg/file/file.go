package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	TreeInfo struct {
		NumFiles  int
		TotalSize int64
	}
)

func (f *Tree) FindMatch(filePath string) (*Tree, error) {
	if filePath == "/" {
		return f, nil
	}
	match := f
	parts := strings.Split(strings.Trim(filePath, "/"), "/")
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
			return nil, fmt.Errorf("%w: %s", ErrNotFound, filePath)
		}
	}
	return match, nil
}

func GetFileTree(filePath string, skipDotFiles bool, errWriter io.Writer) (*Tree, *TreeInfo, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, nil, err
	}
	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, nil, err
	}
	root := &Tree{
		Path:  absPath,
		Name:  stat.Name(),
		Size:  stat.Size(),
		IsDir: stat.IsDir(),
	}
	numFiles := 0
	totalSize := int64(0)
	queue := []*Tree{root}
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
				queue = append(queue, child)
				f.Children = append(f.Children, child)
			}
		} else {
			totalSize += f.Size
		}
		numFiles++
	}
	info := &TreeInfo{
		NumFiles:  numFiles,
		TotalSize: totalSize,
	}
	return root, info, nil
}
