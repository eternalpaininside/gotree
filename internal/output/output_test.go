package output

import (
	"bytes"
	"encoding/json"
	"iter"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eternalpaininside/gotree.git/internal/walker"
)

func testEntries() []walker.Entry {
	return []walker.Entry{
		{
			Path:   "/project",
			Name:   "project",
			Depth:  0,
			Type:   walker.TypeDir,
			IsLast: true,
		},
		{
			Path:   "/project/cmd",
			Name:   "cmd",
			Parent: "/project",
			Depth:  1,
			Type:   walker.TypeDir,
			IsLast: false,
		},
		{
			Path:   "/project/cmd/main.go",
			Name:   "main.go",
			Parent: "/project/cmd",
			Depth:  2,
			Type:   walker.TypeFile,
			IsLast: true,
		},
		{
			Path:   "/project/README.md",
			Name:   "README.md",
			Parent: "/project",
			Depth:  1,
			Type:   walker.TypeFile,
			IsLast: true,
		},
	}
}

func sequence(
	entries []walker.Entry,
) iter.Seq2[walker.Entry, error] {
	return func(yield func(walker.Entry, error) bool) {
		for _, entry := range entries {
			if !yield(entry, nil) {
				return
			}
		}
	}
}

func TestTerminal(t *testing.T) {
	var buf bytes.Buffer

	err := Terminal(&buf, sequence(testEntries()))
	if err != nil {
		t.Fatal(err)
	}

	expected := strings.Join([]string{
		"project",
		"├── cmd/",
		"│   └── main.go",
		"└── README.md",
		"",
	}, "\n")

	if buf.String() != expected {
		t.Errorf(
			"unexpected tree:\ngot:\n%s\nwant:\n%s",
			buf.String(),
			expected,
		)
	}
}

func TestPipeline(t *testing.T) {
	var buf bytes.Buffer

	err := Pipeline(&buf, sequence(testEntries()))
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(
		strings.TrimSpace(buf.String()),
		"\n",
	)

	if len(lines) != 4 {
		t.Fatalf("got %d records, want 4", len(lines))
	}

	for i, line := range lines {
		var record PipelineEntry

		if err := json.Unmarshal(
			[]byte(line),
			&record,
		); err != nil {
			t.Fatalf(
				"invalid JSONL record %d: %v",
				i, err,
			)
		}
	}
}

func TestJSON(t *testing.T) {
	for _, withStats := range []bool{false, true} {
		t.Run(
			map[bool]string{
				false: "without stats",
				true:  "with stats",
			}[withStats],
			func(t *testing.T) {
				var buf bytes.Buffer

				err := JSON(
					&buf,
					sequence(testEntries()),
					withStats,
				)
				if err != nil {
					t.Fatal(err)
				}

				var result struct {
					Entries []PipelineEntry `json:"entries"`
					Stats   *struct {
						Files       uint64 `json:"files"`
						Directories uint64 `json:"directories"`
					} `json:"stats"`
				}

				if err := json.Unmarshal(
					buf.Bytes(),
					&result,
				); err != nil {
					t.Fatal(err)
				}

				if len(result.Entries) != 4 {
					t.Errorf(
						"got %d entries, want 4",
						len(result.Entries),
					)
				}

				if withStats {
					if result.Stats == nil {
						t.Fatal("missing stats")
					}

					if result.Stats.Files != 2 ||
						result.Stats.Directories != 2 {
						t.Errorf(
							"unexpected stats: %+v",
							result.Stats,
						)
					}
				} else if result.Stats != nil {
					t.Error("unexpected stats")
				}
			},
		)
	}
}

func TestPaths(t *testing.T) {
	root := filepath.FromSlash("/project")

	var buf bytes.Buffer

	err := Paths(
		&buf,
		root,
		sequence(testEntries()),
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	expected := strings.Join([]string{
		".",
		"cmd",
		"cmd/main.go",
		"README.md",
		"",
	}, "\n")

	if buf.String() != expected {
		t.Errorf(
			"got:\n%s\nwant:\n%s",
			buf.String(),
			expected,
		)
	}
}

func TestPathsWithStats(t *testing.T) {
	var buf bytes.Buffer

	err := Paths(
		&buf,
		filepath.FromSlash("/project"),
		sequence(testEntries()),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		buf.String(),
		"# files=2 directories=2",
	) {
		t.Errorf(
			"missing statistics:\n%s",
			buf.String(),
		)
	}
}
