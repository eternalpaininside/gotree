package walker

import (
	"iter"
)

// Stats содержит статистику видимых объектов.
type Stats struct {
	Files       uint64 `json:"files"`
	Directories uint64 `json:"directories"`
	Symlinks    uint64 `json:"symlinks"`
	Other       uint64 `json:"other"`
}

// Total возвращает общее количество объектов,
// включая корневой элемент.
func (s Stats) Total() uint64 {
	return s.Files +
		s.Directories +
		s.Symlinks +
		s.Other
}

// Count оборачивает итератор и считает объекты,
// не изменяя последовательность элементов.
//
// Счётчики обновляются при получении каждого
// элемента потребителем.
func Count(
	entries iter.Seq2[Entry, error],
	result *Stats,
) iter.Seq2[Entry, error] {
	return func(yield func(Entry, error) bool) {
		for entry, err := range entries {
			if err != nil {
				yield(entry, err)
				return
			}

			switch entry.Type {
			case TypeFile:
				result.Files++

			case TypeDir:
				result.Directories++

			case TypeSymlink:
				result.Symlinks++

			default:
				result.Other++
			}

			if !yield(entry, nil) {
				return
			}
		}
	}
}
