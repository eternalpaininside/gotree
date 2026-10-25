package walker

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

type EntryType string

const (
	TypeFile    EntryType = "file"
	TypeDir     EntryType = "directory"
	TypeSymlink EntryType = "symlink"
	TypeOther   EntryType = "other"
)

type Entry struct {
	Path   string
	Name   string
	Parent string
	Depth  int
	Type   EntryType
	IsLast bool
}

type Options struct {
	IncludeFiles bool
	ShowAll      bool
}

var ignoredNames = map[string]struct{}{
	"node_modules": {},
	"__pycache__":  {},
	"venv":         {},
	"dist":         {},
	"build":        {},
	"coverage":     {},
	"vendor":       {},
}

func shouldSkip(name string, opts Options) bool {
	if opts.ShowAll {
		return false
	}

	if strings.HasPrefix(name, ".") {
		return true
	}

	_, ignored := ignoredNames[name]
	return ignored
}

func Walk(root string, opts Options) iter.Seq2[Entry, error] {
	return func(yield func(Entry, error) bool) {
		root, err := filepath.Abs(root)
		if err != nil {
			yield(Entry{}, err)
			return
		}

		var visit func(string, string, int, bool) bool

		visit = func(
			path string,
			parent string,
			depth int,
			isLast bool,
		) bool {
			info, err := os.Lstat(path)
			if err != nil {
				yield(Entry{}, err)
				return false
			}

			entryType := classifyInfo(info)

			entry := Entry{
				Path:   path,
				Name:   info.Name(),
				Parent: parent,
				Depth:  depth,
				Type:   entryType,
				IsLast: isLast,
			}

			if opts.IncludeFiles || entryType == TypeDir {
				if !yield(entry, nil) {
					return false
				}
			}

			if entryType != TypeDir {
				return true
			}

			children, err := os.ReadDir(path)
			if err != nil {
				yield(Entry{}, err)
				return false
			}

			filtered := children[:0]

			for _, child := range children {
				if shouldSkip(child.Name(), opts) {
					continue
				}

				if !opts.IncludeFiles && !child.IsDir() {
					continue
				}

				filtered = append(filtered, child)
			}

			for i, child := range filtered {
				childPath := filepath.Join(path, child.Name())

				if !visit(
					childPath,
					path,
					depth+1,
					i == len(filtered)-1,
				) {
					return false
				}
			}

			return true
		}

		visit(root, "", 0, true)
	}
}

func classifyInfo(info fs.FileInfo) EntryType {
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return TypeSymlink
	case info.IsDir():
		return TypeDir
	case info.Mode().IsRegular():
		return TypeFile
	default:
		return TypeOther
	}
}
