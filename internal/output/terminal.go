package output

import (
	"fmt"
	"io"
	"iter"
	"strings"

	"github.com/eternalpaininside/gotree.git/internal/walker"
)

func Terminal(
	w io.Writer,
	entries iter.Seq2[walker.Entry, error],
) error {
	// true означает, что у предка есть следующие соседи.
	branches := make([]bool, 0, 16)

	for entry, err := range entries {
		if err != nil {
			return err
		}

		if entry.Depth == 0 {
			if _, err := fmt.Fprintln(w, entry.Name); err != nil {
				return err
			}
			continue
		}

		// Убираем информацию об уровнях,
		// из которых уже вышли.
		if len(branches) >= entry.Depth {
			branches = branches[:entry.Depth-1]
		}

		var b strings.Builder

		for _, hasNext := range branches {
			if hasNext {
				b.WriteString("│   ")
			} else {
				b.WriteString("    ")
			}
		}

		if entry.IsLast {
			b.WriteString("└── ")
		} else {
			b.WriteString("├── ")
		}

		b.WriteString(entry.Name)

		if entry.Type == walker.TypeDir {
			b.WriteString("/")
		}

		if _, err := fmt.Fprintln(w, b.String()); err != nil {
			return err
		}

		branches = append(branches, !entry.IsLast)
	}

	return nil
}
