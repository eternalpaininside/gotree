package output

import (
	"fmt"
	"io"
	"iter"
	"path/filepath"

	"github.com/eternalpaininside/gotree.git/internal/stats"
	"github.com/eternalpaininside/gotree.git/internal/walker"
)

// Paths выводит относительные пути по одному на строку.
//
// root - корень обхода файловой системы.
// entries - поток объектов от walker.
// includeStats - добавлять ли статистику в stdout.
//
// Полное дерево в памяти не хранится.
func Paths(
	w io.Writer,
	root string,
	entries iter.Seq2[walker.Entry, error],
	includeStats bool,
) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	var result stats.Stats

	if includeStats {
		entries = stats.Count(entries, &result)
	}

	for entry, err := range entries {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(absRoot, entry.Path)
		if err != nil {
			return fmt.Errorf(
				"get relative path for %q: %w",
				entry.Path,
				err,
			)
		}

		// Используем прямые слеши независимо
		// от операционной системы.
		relativePath = filepath.ToSlash(relativePath)

		if _, err := fmt.Fprintln(w, relativePath); err != nil {
			return fmt.Errorf(
				"write path %q: %w",
				relativePath,
				err,
			)
		}
	}

	if includeStats {
		if _, err := fmt.Fprintf(
			w,
			"# files=%d directories=%d symlinks=%d other=%d total=%d\n",
			result.Files,
			result.Directories,
			result.Symlinks,
			result.Other,
			result.Total(),
		); err != nil {
			return fmt.Errorf("write statistics: %w", err)
		}
	}

	return nil
}
