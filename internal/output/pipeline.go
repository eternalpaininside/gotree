package output

import (
	"encoding/json"
	"io"
	"iter"

	"github.com/eternalpaininside/gotree.git/internal/walker"
)

type PipelineEntry struct {
	Path   string           `json:"path"`
	Name   string           `json:"name"`
	Parent string           `json:"parent,omitempty"`
	Depth  int              `json:"depth"`
	Type   walker.EntryType `json:"type"`
	IsLast bool             `json:"is_last"`
}

func Pipeline(
	w io.Writer,
	entries iter.Seq2[walker.Entry, error],
) error {
	encoder := json.NewEncoder(w)

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

		if err := encoder.Encode(record); err != nil {
			return err
		}
	}

	return nil
}
