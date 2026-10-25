package filter

import "testing"

func TestDefaultIgnore(t *testing.T) {
	f, err := New(FilterOptions{})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path   string
		isDir  bool
		ignore bool
	}{
		{".git", true, true},
		{".idea/workspace.xml", false, true},
		{".DS_Store", false, true},
		{".env", false, true},
		{"node_modules", true, true},
		{"__pycache__", true, true},
		{"dist", true, true},
		{"build", true, true},
		{"vendor", true, false},
		{"go.mod", false, false},
		{"main.go", false, false},
		{"internal/service.go", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := f.Ignore(tt.path, tt.isDir)

			if got != tt.ignore {
				t.Errorf(
					"Ignore(%q) = %v, want %v",
					tt.path, got, tt.ignore,
				)
			}
		})
	}
}

func TestExcludePatterns(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		isDir    bool
		expected bool
	}{
		{
			"extension",
			"*.log",
			"logs/app.log",
			false,
			true,
		},
		{
			"extension mismatch",
			"*.log",
			"logs/app.txt",
			false,
			false,
		},
		{
			"directory",
			"tmp/",
			"internal/tmp",
			true,
			true,
		},
		{
			"directory only",
			"tmp/",
			"tmp",
			false,
			false,
		},
		{
			"exact relative path",
			"internal/cache",
			"internal/cache",
			true,
			true,
		},
		{
			"recursive wildcard",
			"**/cache",
			"internal/service/cache",
			true,
			true,
		},
		{
			"recursive wildcard at root",
			"**/cache",
			"cache",
			true,
			true,
		},
		{
			"recursive contents",
			"**/cache/**",
			"internal/cache/data.bin",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := New(FilterOptions{
				Exclude: []string{tt.pattern},
			})
			if err != nil {
				t.Fatal(err)
			}

			got := f.Ignore(tt.path, tt.isDir)

			if got != tt.expected {
				t.Errorf(
					"pattern=%q path=%q: got %v, want %v",
					tt.pattern,
					tt.path,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestShowAllOverridesFilters(t *testing.T) {
	f, err := New(FilterOptions{
		ShowAll: true,
		Exclude: []string{
			"*.go",
			"vendor/",
			"tmp/",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		".git/config",
		".DS_Store",
		"main.go",
		"vendor",
		"tmp",
	} {
		if f.Ignore(path, true) {
			t.Errorf(
				"ShowAll must allow %q",
				path,
			)
		}
	}
}

func TestInvalidPatterns(t *testing.T) {
	for _, pattern := range []string{
		"",
		"[invalid",
		"src/[abc",
	} {
		t.Run(pattern, func(t *testing.T) {
			_, err := New(FilterOptions{
				Exclude: []string{pattern},
			})

			if err == nil {
				t.Errorf(
					"expected error for %q",
					pattern,
				)
			}
		})
	}
}
