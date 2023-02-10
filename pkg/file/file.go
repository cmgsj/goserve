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
	Size     string
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

func GetFileTree(fpath string, skipDotFiles bool, errWriter io.Writer) (*FileTree, int, error) {
	abspath, err := filepath.Abs(fpath)
	if err != nil {
		return nil, 0, err
	}
	fstat, err := os.Stat(abspath)
	if err != nil {
		return nil, 0, err
	}
	root := &FileTree{
		Path:  abspath,
		Name:  fstat.Name(),
		IsDir: fstat.IsDir(),
	}
	numfiles := 0
	queue := []*FileTree{root}
	for len(queue) > 0 {
		f := queue[0]
		queue = queue[1:]
		numfiles++
		if f.IsDir {
			f.Size = " - "
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
				f.Children = append(f.Children, child)
				queue = append(queue, child)
			}
		} else {
			f.Size = FormatSize(fstat.Size())
		}
	}
	return root, numfiles, nil
}

func FormatSize(fsize int64) string {
	var (
		unit   string
		factor int64
	)
	if factor = 1024 * 1024 * 2014; fsize > factor {
		unit = "GB"
	} else if factor = 1024 * 1024; fsize > factor {
		unit = "MB"
	} else if factor = 1024; fsize > factor {
		unit = "KB"
	} else {
		unit = "B"
		factor = 1
	}
	return fmt.Sprintf("%.2f%s", float64(fsize)/float64(factor), unit)
}
