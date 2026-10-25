package walker

import (
	"fmt"
	"io/fs"
	"iter"
	"os"
	"path/filepath"

	"github.com/eternalpaininside/gotree.git/internal/walker/filter"
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

// Walk возвращает поток элементов файловой системы.
//
// Каждый элемент передаётся через yield.
// Полное дерево в памяти не хранится.
//
// При ошибке обход прекращается, а ошибка
// передаётся потребителю через итератор.
func Walk(
	root string,
	opts Options,
) iter.Seq2[Entry, error] {
	return func(yield func(Entry, error) bool) {
		if err := opts.Validate(); err != nil {
			yield(Entry{}, err)
			return
		}

		absRoot, err := filepath.Abs(root)
		if err != nil {
			yield(Entry{}, err)
			return
		}

		fileFilter, err := filter.New(filter.FilterOptions{
			ShowAll: opts.ShowAll,
			Exclude: opts.Exclude,
		})
		if err != nil {
			yield(Entry{}, err)
			return
		}

		var gitFilter *filter.GitIgnore

		if opts.UseGitIgnore && !opts.ShowAll {
			gitFilter, err = filter.NewGitIgnore(absRoot)
			if err != nil {
				yield(Entry{}, err)
				return
			}
		}

		// visit возвращает false, если обход
		// нужно остановить.
		var visit func(
			path string,
			parent string,
			depth int,
			isLast bool,
		) bool

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

			// Передаём текущий элемент потребителю.
			if opts.IncludeFiles || entryType == TypeDir {
				if !yield(entry, nil) {
					return false
				}
			}

			// Файлы и символические ссылки
			// не имеют раскрываемых потомков.
			if entryType != TypeDir {
				return true
			}

			// Не заходим глубже заданного уровня.
			if opts.MaxDepth >= 0 &&
				depth >= opts.MaxDepth {
				return true
			}

			// Загружаем правила .gitignore текущей
			// директории перед чтением её детей.
			if gitFilter != nil {
				count, err := gitFilter.Push(path)
				if err != nil {
					yield(Entry{}, err)
					return false
				}

				defer gitFilter.Pop(count)
			}

			children, err := os.ReadDir(path)
			if err != nil {
				yield(Entry{}, fmt.Errorf(
					"read directory %s: %w",
					path,
					err,
				))
				return false
			}

			// Отбираем только видимых потомков.
			// Исходный children можно использовать
			// как буфер без новой аллокации.
			visible := children[:0]

			for _, child := range children {
				childPath := filepath.Join(
					path,
					child.Name(),
				)

				relativePath, err := filepath.Rel(
					absRoot,
					childPath,
				)
				if err != nil {
					yield(Entry{}, err)
					return false
				}

				relativePath = filepath.ToSlash(
					relativePath,
				)

				isDir := child.IsDir()

				if fileFilter.Ignore(
					relativePath,
					isDir,
				) {
					continue
				}

				if gitFilter != nil &&
					gitFilter.Ignore(relativePath, isDir) {
					continue
				}

				if !opts.IncludeFiles && !isDir {
					continue
				}

				visible = append(visible, child)
			}

			// Передаём потомков последовательно.
			for i, child := range visible {
				childPath := filepath.Join(
					path,
					child.Name(),
				)

				if !visit(
					childPath,
					path,
					depth+1,
					i == len(visible)-1,
				) {
					return false
				}
			}

			return true
		}

		visit(absRoot, "", 0, true)
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
