package filter

import (
	"os"
	"path/filepath"
	"testing"
)

func writeGitIgnore(
	t *testing.T,
	root, dir, content string,
) {
	t.Helper()

	target := filepath.Join(root, dir)

	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(target, ".gitignore"),
		[]byte(content),
		0644,
	); err != nil {
		t.Fatal(err)
	}
}

func TestGitIgnoreRootRules(t *testing.T) {
	root := t.TempDir()

	writeGitIgnore(
		t, root, ".",
		"*.log\nbuild/\n!important.log\n",
	)

	g, err := NewGitIgnore(root)
	if err != nil {
		t.Fatal(err)
	}

	count, err := g.Push(root)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Pop(count)

	tests := []struct {
		path   string
		isDir  bool
		ignore bool
	}{
		{"debug.log", false, true},
		{"src/debug.log", false, true},
		{"important.log", false, false},
		{"src/important.log", false, false},
		{"build", true, true},
		{"src/build", true, true},
		{"main.go", false, false},
	}

	for _, tt := range tests {
		if got := g.Ignore(tt.path, tt.isDir); got != tt.ignore {
			t.Errorf(
				"Ignore(%q) = %v, want %v",
				tt.path, got, tt.ignore,
			)
		}
	}
}

func TestGitIgnoreNestedRules(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")

	writeGitIgnore(t, root, ".", "*.log\n")
	writeGitIgnore(t, root, "src", "!important.log\n")

	g, err := NewGitIgnore(root)
	if err != nil {
		t.Fatal(err)
	}

	rootCount, err := g.Push(root)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Pop(rootCount)

	if !g.Ignore("src/important.log", false) {
		t.Fatal("root rule should ignore important.log")
	}

	srcCount, err := g.Push(src)
	if err != nil {
		t.Fatal(err)
	}

	if g.Ignore("src/important.log", false) {
		t.Error("nested negation should restore important.log")
	}

	if !g.Ignore("src/debug.log", false) {
		t.Error("debug.log should remain ignored")
	}

	g.Pop(srcCount)

	if !g.Ignore("src/important.log", false) {
		t.Error("nested rules leaked after Pop")
	}
}

func TestGitIgnoreMissingFile(t *testing.T) {
	root := t.TempDir()

	g, err := NewGitIgnore(root)
	if err != nil {
		t.Fatal(err)
	}

	count, err := g.Push(root)
	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Errorf("got %d patterns, want 0", count)
	}

	g.Pop(count)
}

func TestGitIgnoreDirectoryScope(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")

	writeGitIgnore(t, root, "src", "private.txt\n")

	g, err := NewGitIgnore(root)
	if err != nil {
		t.Fatal(err)
	}

	count, err := g.Push(src)
	if err != nil {
		t.Fatal(err)
	}
	defer g.Pop(count)

	if !g.Ignore("src/private.txt", false) {
		t.Error("file inside src should be ignored")
	}

	if g.Ignore("other/private.txt", false) {
		t.Error("src rules must not affect other directories")
	}
}
