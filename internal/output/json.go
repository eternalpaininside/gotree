package output

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"

	"github.com/eternalpaininside/gotree.git/internal/walker"
)

// JSON записывает поток элементов как единый JSON-объект.
//
// При includeStats = true в результат добавляется
// статистика объектов.
//
// Функция не накапливает элементы в памяти.
func JSON(
	w io.Writer,
	entries iter.Seq2[walker.Entry, error],
	includeStats bool,
) error {
	if _, err := io.WriteString(w, "{\n  \"entries\": ["); err != nil {
		return err
	}

	var result walker.Stats
	first := true

	for entry, err := range entries {
		if err != nil {
			return err
		}

		record := PipelineEntry{
			Path:   entry.Path,
			Name:   entry.Name,
			Parent: entry.Parent,
			Depth:  entry.Depth,
			Type:   entry.Type,
			IsLast: entry.IsLast,
		}

		data, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf(
				"marshal entry %q: %w",
				entry.Path,
				err,
			)
		}

		if !first {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		}

		if _, err := io.WriteString(w, "\n    "); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}

		first = false

		// Считаем только успешно сериализованные
		// и записанные элементы.
		switch entry.Type {
		case walker.TypeFile:
			result.Files++

		case walker.TypeDir:
			result.Directories++

		case walker.TypeSymlink:
			result.Symlinks++

		default:
			result.Other++
		}
	}

	if _, err := io.WriteString(w, "\n  ]"); err != nil {
		return err
	}

	if includeStats {
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(
			w,
			",\n  \"stats\": ",
		); err != nil {
			return err
		}

		if _, err := w.Write(data); err != nil {
			return err
		}
	}

	_, err := io.WriteString(w, "\n}\n")
	return err
}
