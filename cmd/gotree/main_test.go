package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func cliFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	if err := os.MkdirAll(
		filepath.Join(root, ".git"),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	for name, content := range map[string]string{
		"main.go":    "package main",
		"README.md":  "# test",
		"debug.log":  "debug",
		".DS_Store":  "metadata",
		".gitignore": "*.log\n",
	} {
		if err := os.WriteFile(
			filepath.Join(root, name),
			[]byte(content),
			0644,
		); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command("go", append(
		[]string{"run", "."},
		args...,
	)...)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestCLITree(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-f", "tree", root,
	)
	if err != nil {
		t.Fatal(err, output)
	}

	if !strings.Contains(output, "main.go") {
		t.Error("missing main.go")
	}

	if strings.Contains(output, ".git/") {
		t.Error(".git must be hidden by default")
	}

	if strings.Contains(output, ".DS_Store") {
		t.Error(".DS_Store must be hidden")
	}
}

func TestCLIShowAll(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-a", "-f", "tree", root,
	)
	if err != nil {
		t.Fatal(err, output)
	}

	if !strings.Contains(output, ".git/") {
		t.Error("-a must show .git")
	}

	if !strings.Contains(output, ".DS_Store") {
		t.Error("-a must show .DS_Store")
	}
}

func TestCLIGitIgnore(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-g", "-f", "tree", root,
	)
	if err != nil {
		t.Fatal(err, output)
	}

	if strings.Contains(output, "debug.log") {
		t.Error("-g must exclude debug.log")
	}

	if !strings.Contains(output, "main.go") {
		t.Error("main.go must remain visible")
	}
}

func TestCLIJSON(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-f", "json", "-s", root,
	)
	if err != nil {
		t.Fatal(err, output)
	}

	var result struct {
		Entries []json.RawMessage `json:"entries"`
		Stats   json.RawMessage   `json:"stats"`
	}

	if err := json.Unmarshal(
		[]byte(output),
		&result,
	); err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) == 0 {
		t.Error("empty entries")
	}

	if len(result.Stats) == 0 {
		t.Error("missing stats")
	}
}

func TestCLIInvalidDepth(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-L", "-2", root,
	)

	if err == nil {
		t.Fatal("expected nonzero exit code")
	}

	if !strings.Contains(output, "invalid max depth") {
		t.Errorf("unexpected error: %s", output)
	}
}

func TestCLIInvalidFormat(t *testing.T) {
	root := cliFixture(t)

	output, err := runCLI(
		t, "-f", "xml", root,
	)

	if err == nil {
		t.Fatal("expected nonzero exit code")
	}

	if !strings.Contains(output, "unknown format") {
		t.Errorf("unexpected error: %s", output)
	}
}
