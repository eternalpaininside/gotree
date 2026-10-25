package walker

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func createTree(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	dirs := []string{
		"cmd",
		"internal",
		"internal/service",
		".git",
		"node_modules",
		"vendor",
	}

	files := []string{
		"cmd/main.go",
		"internal/service/service.go",
		"README.md",
		".git/config",
		".DS_Store",
		"node_modules/index.js",
		"vendor/dependency.go",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(
			filepath.Join(root, dir),
			0755,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range files {
		err := os.WriteFile(
			filepath.Join(root, file),
			[]byte("test"),
			0644,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func collect(
	t *testing.T,
	root string,
	opts Options,
) []Entry {
	t.Helper()

	var result []Entry

	for entry, err := range Walk(root, opts) {
		if err != nil {
			t.Fatal(err)
		}

		result = append(result, entry)
	}

	return result
}

func names(entries []Entry) []string {
	result := make([]string, 0, len(entries))

	for _, entry := range entries {
		result = append(result, entry.Name)
	}

	return result
}

func TestWalkDefault(t *testing.T) {
	root := createTree(t)
	opts := DefaultOptions()

	entries := collect(t, root, opts)

	expected := []string{
		filepath.Base(root),
		"README.md",
		"cmd",
		"main.go",
		"internal",
		"service",
		"service.go",
		"vendor",
		"dependency.go",
	}

	if !reflect.DeepEqual(names(entries), expected) {
		t.Errorf(
			"unexpected entries:\ngot:  %v\nwant: %v",
			names(entries),
			expected,
		)
	}
}

func TestWalkShowAll(t *testing.T) {
	root := createTree(t)

	opts := DefaultOptions()
	opts.ShowAll = true
	opts.Exclude = []string{"*.go"}

	entries := collect(t, root, opts)

	found := make(map[string]bool)

	for _, entry := range entries {
		found[entry.Name] = true
	}

	for _, name := range []string{
		".git",
		".DS_Store",
		"node_modules",
		"main.go",
	} {
		if !found[name] {
			t.Errorf(
				"expected %q with ShowAll",
				name,
			)
		}
	}
}

func TestWalkDirectoriesOnly(t *testing.T) {
	root := createTree(t)

	opts := DefaultOptions()
	opts.IncludeFiles = false

	entries := collect(t, root, opts)

	for _, entry := range entries {
		if entry.Type != TypeDir {
			t.Errorf(
				"unexpected non-directory: %s",
				entry.Path,
			)
		}
	}
}

func TestWalkMaxDepth(t *testing.T) {
	root := createTree(t)

	tests := []struct {
		name     string
		maxDepth int
		expected int
	}{
		{"root only", 0, 1},
		{"first level", 1, 5},
		{"second level", 2, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.MaxDepth = tt.maxDepth

			entries := collect(t, root, opts)

			if len(entries) != tt.expected {
				t.Errorf(
					"got %d entries, want %d",
					len(entries),
					tt.expected,
				)
			}

			for _, entry := range entries {
				if entry.Depth > tt.maxDepth {
					t.Errorf(
						"entry %s exceeds max depth",
						entry.Path,
					)
				}
			}
		})
	}
}

func TestWalkIsLast(t *testing.T) {
	root := createTree(t)

	opts := DefaultOptions()
	opts.IncludeFiles = false

	entries := collect(t, root, opts)

	// Для каждого родителя проверяем,
	// что последний видимый потомок имеет IsLast.
	lastByParent := make(map[string]Entry)

	for _, entry := range entries {
		if entry.Depth == 0 {
			continue
		}

		lastByParent[entry.Parent] = entry
	}

	for _, entry := range lastByParent {
		if !entry.IsLast {
			t.Errorf(
				"last child %q has IsLast=false",
				entry.Path,
			)
		}
	}

	// У каждого родителя должен быть ровно
	// один видимый потомок с IsLast=true.
	counts := make(map[string]int)

	for _, entry := range entries {
		if entry.Depth > 0 && entry.IsLast {
			counts[entry.Parent]++
		}
	}

	for parent := range lastByParent {
		if counts[parent] != 1 {
			t.Errorf(
				"parent %q has %d last children",
				parent,
				counts[parent],
			)
		}
	}
}

func TestWalkExclude(t *testing.T) {
	root := createTree(t)

	opts := DefaultOptions()
	opts.Exclude = []string{
		"*.go",
		"vendor/",
	}

	entries := collect(t, root, opts)

	for _, entry := range entries {
		if filepath.Ext(entry.Name) == ".go" {
			t.Errorf(
				"excluded Go file appeared: %s",
				entry.Path,
			)
		}

		if entry.Name == "vendor" {
			t.Error("excluded vendor directory appeared")
		}
	}
}

func TestWalkEarlyStop(t *testing.T) {
	root := createTree(t)

	opts := DefaultOptions()

	count := 0

	for _, err := range Walk(root, opts) {
		if err != nil {
			t.Fatal(err)
		}

		count++

		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf(
			"expected 2 entries, got %d",
			count,
		)
	}
}

func TestWalkInvalidDepth(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxDepth = -2

	for _, err := range Walk(".", opts) {
		if err == nil {
			t.Fatal("expected validation error")
		}
		return
	}

	t.Fatal("expected iterator to return an error")
}

func TestWalkSymlink(t *testing.T) {
	root := t.TempDir()

	target := filepath.Join(root, "target.txt")
	link := filepath.Join(root, "link.txt")

	if err := os.WriteFile(
		target,
		[]byte("test"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	entries := collect(t, root, DefaultOptions())

	for _, entry := range entries {
		if entry.Name == "link.txt" {
			if entry.Type != TypeSymlink {
				t.Errorf(
					"expected symlink, got %s",
					entry.Type,
				)
			}
			return
		}
	}

	t.Fatal("symlink not found")
}
