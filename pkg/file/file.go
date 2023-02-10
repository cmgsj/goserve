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
	ErrFileNotFound = errors.New("file not found")
)

type FileTree struct {
	Path     string
	Name     string
	Size     int64
	IsDir    bool
	IsBadDir bool
	Children []*FileTree
}

func (f *FileTree) FindMatch(fpath string) (*FileTree, error) {
	if fpath == "/" {
		return f, nil
	}
	parts := strings.Split(strings.Trim(fpath, "/"), "/")
	for _, part := range parts {
		found := false
		for _, child := range f.Children {
			if child.Name == part {
				f = child
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, fpath)
		}
	}
	return f, nil
}

func GetFileTree(fpath string, skipDotFiles bool, errWriter io.Writer) (*FileTree, int, int64, error) {
	abspath, err := filepath.Abs(fpath)
	if err != nil {
		return nil, 0, 0, err
	}
	fstat, err := os.Stat(abspath)
	if err != nil {
		return nil, 0, 0, err
	}
	root := &FileTree{
		Path:  abspath,
		Name:  fstat.Name(),
		IsDir: fstat.IsDir(),
	}
	numfiles := 0
	totalSize := int64(0)
	queue := []*FileTree{root}
	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]
		if f.IsDir {
			entries, err := os.ReadDir(f.Path)
			if err != nil {
				f.IsBadDir = true
				fmt.Fprintln(errWriter, err)
				continue
			}
			for _, entry := range entries {
				if skipDotFiles && strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				finfo, err := entry.Info()
				if err != nil {
					fmt.Fprintln(errWriter, err)
					continue
				}
				child := &FileTree{
					Path:  filepath.Join(f.Path, entry.Name()),
					Name:  entry.Name(),
					IsDir: finfo.IsDir(),
				}
				queue = append(queue, child)
				f.Children = append(f.Children, child)
			}
		} else {
			f.Size = fstat.Size()
			totalSize += f.Size
		}
		numfiles++
	}
	return root, numfiles, totalSize, nil
}
