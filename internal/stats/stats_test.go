package stats

import (
	"iter"
	"testing"

	"github.com/eternalpaininside/gotree.git/internal/walker"
)

func entries() iter.Seq2[walker.Entry, error] {
	data := []walker.Entry{
		{Type: walker.TypeDir},
		{Type: walker.TypeDir},
		{Type: walker.TypeFile},
		{Type: walker.TypeFile},
		{Type: walker.TypeFile},
		{Type: walker.TypeSymlink},
		{Type: walker.TypeOther},
	}

	return func(yield func(walker.Entry, error) bool) {
		for _, entry := range data {
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func TestCount(t *testing.T) {
	var result Stats

	for _, err := range Count(entries(), &result) {
		if err != nil {
			t.Fatal(err)
		}
	}

	if result.Files != 3 {
		t.Errorf("files: got %d, want 3", result.Files)
	}

	if result.Directories != 2 {
		t.Errorf(
			"directories: got %d, want 2",
			result.Directories,
		)
	}

	if result.Symlinks != 1 {
		t.Errorf(
			"symlinks: got %d, want 1",
			result.Symlinks,
		)
	}

	if result.Other != 1 {
		t.Errorf("other: got %d, want 1", result.Other)
	}

	if result.Total() != 7 {
		t.Errorf("total: got %d, want 7", result.Total())
	}
}

func TestCountEarlyStop(t *testing.T) {
	var result Stats

	processed := 0

	for _, err := range Count(entries(), &result) {
		if err != nil {
			t.Fatal(err)
		}

		processed++

		if processed == 3 {
			break
		}
	}

	if result.Total() != 3 {
		t.Errorf(
			"got %d counted objects, want 3",
			result.Total(),
		)
	}
}

func TestEmptyStats(t *testing.T) {
	var result Stats

	if result.Total() != 0 {
		t.Errorf(
			"empty stats total = %d",
			result.Total(),
		)
	}
}
