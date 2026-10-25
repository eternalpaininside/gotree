package filter

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

var defaultIgnored = map[string]struct{}{
	"node_modules": {},
	"__pycache__":  {},
	"venv":         {},
	".venv":        {},
	"dist":         {},
	"build":        {},
	"coverage":     {},
}

type FilterOptions struct {
	ShowAll bool
	Exclude []string
}

// Filter проверяет, должен ли объект быть исключён.
type Filter struct {
	opts FilterOptions
}

// New создаёт фильтр и проверяет шаблоны.
func New(opts FilterOptions) (*Filter, error) {
	for _, pattern := range opts.Exclude {
		if err := validatePattern(pattern); err != nil {
			return nil, fmt.Errorf(
				"invalid exclude pattern %q: %w",
				pattern,
				err,
			)
		}
	}

	return &Filter{opts: opts}, nil
}

// Ignore возвращает true, если объект нужно скрыть.
//
// relativePath задаётся относительно корня обхода
// с разделителем '/', например:
//
//	internal/cache/data.log
//
// isDir указывает, является ли объект директорией.
func (f *Filter) Ignore(
	relativePath string,
	isDir bool,
) bool {
	if f.opts.ShowAll {
		return false
	}

	relativePath = strings.TrimPrefix(
		relativePath,
		"./",
	)

	relativePath = strings.Trim(
		relativePath,
		"/",
	)

	if relativePath == "" || relativePath == "." {
		return false
	}

	parts := strings.Split(relativePath, "/")

	// Скрываем любой объект, содержащий скрытый
	// компонент пути.
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}

		if _, exists := defaultIgnored[part]; exists {
			return true
		}
	}

	for _, pattern := range f.opts.Exclude {
		if matchPattern(pattern, relativePath, isDir) {
			return true
		}
	}

	return false
}

// matchPattern проверяет пользовательский шаблон.
//
// Поддерживаются:
//
//	*.log
//	tmp/
//	internal/cache
//	**/cache/**
func matchPattern(
	pattern string,
	relativePath string,
	isDir bool,
) bool {
	directoryOnly := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")

	if directoryOnly && !isDir {
		return false
	}

	// Шаблон без '/' применяется к имени
	// объекта на любой глубине.
	if !strings.Contains(pattern, "/") {
		name := path.Base(relativePath)

		matched, _ := path.Match(pattern, name)
		return matched
	}

	// Шаблон с '/' проверяется относительно
	// корня обхода.
	return matchSegments(
		strings.Split(pattern, "/"),
		strings.Split(relativePath, "/"),
	)
}

// matchSegments реализует поддержку **,
// соответствующего нулю или более компонентам пути.
func matchSegments(pattern, target []string) bool {
	if len(pattern) == 0 {
		return len(target) == 0
	}

	if pattern[0] == "**" {
		// ** соответствует нулю компонентов.
		if matchSegments(pattern[1:], target) {
			return true
		}

		// Или одному и более компонентам.
		if len(target) > 0 {
			return matchSegments(pattern, target[1:])
		}

		return false
	}

	if len(target) == 0 {
		return false
	}

	matched, err := path.Match(pattern[0], target[0])
	if err != nil || !matched {
		return false
	}

	return matchSegments(pattern[1:], target[1:])
}

func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern")
	}

	pattern = strings.TrimSuffix(pattern, "/")

	for _, segment := range strings.Split(pattern, "/") {
		if segment == "**" {
			continue
		}

		if _, err := path.Match(segment, "test"); err != nil {
			return err
		}
	}

	return nil
}

// IsDirEntry - вспомогательная функция для walker.
func IsDirEntry(entry fs.DirEntry) bool {
	return entry.IsDir()
}
