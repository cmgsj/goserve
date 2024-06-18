package files

import (
	"cmp"
	"slices"
)

const (
	RootDir   = "."
	ParentDir = ".."
)

type File struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Size  string `json:"size,omitempty"`
	IsDir bool   `json:"is_dir"`
}

func Sort(files []File) {
	slices.SortFunc(files, Compare)
}

func Compare(x, y File) int {
	if x.IsDir != y.IsDir {
		if x.IsDir {
			return -1
		}
		return +1
	}
	if x.Name == RootDir || x.Name == ParentDir {
		return -1
	}
	return cmp.Compare(x.Name, y.Name)
}
