package filter

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// GitIgnore хранит правила, обнаруженные при обходе.
//
// Правила из вложенных .gitignore имеют приоритет
// над правилами родительских директорий.
type GitIgnore struct {
	root     string
	patterns []gitignore.Pattern
}

// NewGitIgnore создаёт обработчик правил.
//
// root должен указывать на корень обхода.
func NewGitIgnore(root string) (*GitIgnore, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &GitIgnore{
		root: absRoot,
	}, nil
}

// Push загружает .gitignore из указанной директории.
//
// Возвращает количество добавленных правил.
// Это значение нужно передать в Pop после завершения
// обхода директории.
func (g *GitIgnore) Push(dir string) (int, error) {
	filename := filepath.Join(dir, ".gitignore")

	file, err := os.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf(
			"open %s: %w",
			filename,
			err,
		)
	}
	defer file.Close()

	relativeDir, err := filepath.Rel(g.root, dir)
	if err != nil {
		return 0, err
	}

	var domain []string

	if relativeDir != "." {
		domain = strings.Split(
			filepath.ToSlash(relativeDir),
			"/",
		)
	}

	var patterns []gitignore.Pattern

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// CompileMatchPattern самостоятельно
		// обрабатывает комментарии и отрицания.
		pattern := gitignore.ParsePattern(line, domain)
		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf(
			"read %s: %w",
			filename,
			err,
		)
	}

	g.patterns = append(g.patterns, patterns...)

	return len(patterns), nil
}

// Pop удаляет правила, загруженные для директории,
// после завершения обхода её содержимого.
func (g *GitIgnore) Pop(count int) {
	if count <= 0 {
		return
	}

	g.patterns = g.patterns[:len(g.patterns)-count]
}

// Ignore проверяет, исключён ли объект правилами Git.
//
// relativePath должен быть относительно корня обхода,
// с разделителями '/'.
func (g *GitIgnore) Ignore(
	relativePath string,
	isDir bool,
) bool {
	relativePath = filepath.ToSlash(relativePath)

	parts := strings.Split(relativePath, "/")

	matcher := gitignore.NewMatcher(g.patterns)

	return matcher.Match(parts, isDir)
}
