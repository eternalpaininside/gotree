package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/eternalpaininside/gotree.git/internal/output"
	"github.com/eternalpaininside/gotree.git/internal/stats"
	"github.com/eternalpaininside/gotree.git/internal/walker"
	"golang.org/x/term"
)

// excludeFlags позволяет многократно использовать -e.
type excludeFlags []string

func (e *excludeFlags) String() string {
	return strings.Join(*e, ", ")
}

func (e *excludeFlags) Set(value string) error {
	if value == "" {
		return fmt.Errorf("exclude pattern cannot be empty")
	}

	*e = append(*e, value)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gotree:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dirsOnly  bool
		showAll   bool
		gitIgnore bool
		showStats bool
		maxDepth  int
		format    string
		excludes  excludeFlags
	)

	flag.BoolVar(&dirsOnly, "d", false, "show directories only")
	flag.BoolVar(&dirsOnly, "dirs-only", false, "show directories only")

	flag.BoolVar(&showAll, "a", false, "disable all exclusion filters")
	flag.BoolVar(&showAll, "all", false, "disable all exclusion filters")

	flag.BoolVar(&gitIgnore, "g", false, "apply .gitignore rules")
	flag.BoolVar(&gitIgnore, "gitignore", false, "apply .gitignore rules")

	flag.BoolVar(&showStats, "s", false, "print statistics")
	flag.BoolVar(&showStats, "stats", false, "print statistics")

	flag.IntVar(&maxDepth, "L", -1, "maximum depth (-1 = unlimited)")
	flag.IntVar(&maxDepth, "max-depth", -1, "maximum depth (-1 = unlimited)")

	flag.StringVar(&format, "f", "auto", "output format")
	flag.StringVar(&format, "format", "auto", "output format")

	flag.Var(&excludes, "e", "exclude pattern (repeatable)")
	flag.Var(&excludes, "exclude", "exclude pattern (repeatable)")

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage: gotree [flags] [path]\n\n",
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() > 1 {
		return fmt.Errorf("expected at most one path")
	}

	root := "."
	if flag.NArg() == 1 {
		root = flag.Arg(0)
	}

	if format == "auto" {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			format = "tree"
		} else {
			format = "jsonl"
		}
	}

	opts := walker.Options{
		IncludeFiles: !dirsOnly,
		ShowAll:      showAll,
		MaxDepth:     maxDepth,
		Exclude:      excludes,
		UseGitIgnore: gitIgnore,
	}

	if err := opts.Validate(); err != nil {
		return err
	}

	entries := walker.Walk(root, opts)

	switch format {
	case "tree":
		var result stats.Stats

		if showStats {
			entries = stats.Count(entries, &result)
		}

		if err := output.Terminal(os.Stdout, entries); err != nil {
			return err
		}

		if showStats {
			_, err := fmt.Fprintf(
				os.Stdout,
				"\nFiles: %d, Directories: %d, Symlinks: %d, Other: %d, Total: %d\n",
				result.Files,
				result.Directories,
				result.Symlinks,
				result.Other,
				result.Total(),
			)
			return err
		}

		return nil

	case "jsonl":
		var result stats.Stats

		if showStats {
			entries = stats.Count(entries, &result)
		}

		if err := output.Pipeline(os.Stdout, entries); err != nil {
			return err
		}

		if showStats {
			_, err := fmt.Fprintf(
				os.Stdout,
				"{\"record_type\":\"stats\",\"files\":%d,\"directories\":%d,\"symlinks\":%d,\"other\":%d,\"total\":%d}\n",
				result.Files,
				result.Directories,
				result.Symlinks,
				result.Other,
				result.Total(),
			)
			return err
		}

		return nil

	case "json":
		return output.JSON(os.Stdout, entries, showStats)

	case "paths":
		return output.Paths(
			os.Stdout,
			root,
			entries,
			showStats,
		)

	default:
		return fmt.Errorf(
			"unknown format %q: expected tree, jsonl, json or paths",
			format,
		)
	}
}
