package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/eternalpaininside/gotree.git/internal/output"
	"github.com/eternalpaininside/gotree.git/internal/walker"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gotree:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dirsOnly bool
		showAll  bool
		format   string
	)

	flag.BoolVar(
		&dirsOnly,
		"d",
		false,
		"show directories only",
	)
	flag.BoolVar(
		&dirsOnly,
		"dirs-only",
		false,
		"show directories only",
	)

	flag.BoolVar(
		&showAll,
		"a",
		false,
		"show hidden and ignored files",
	)
	flag.BoolVar(
		&showAll,
		"all",
		false,
		"show hidden and ignored files",
	)

	flag.StringVar(
		&format,
		"f",
		"auto",
		"output format: auto, tree, jsonl",
	)
	flag.StringVar(
		&format,
		"format",
		"auto",
		"output format: auto, tree, jsonl",
	)

	flag.Parse()

	if flag.NArg() > 1 {
		return fmt.Errorf(
			"expected at most one path",
		)
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
	}

	entries := walker.Walk(root, opts)

	switch format {
	case "tree":
		return output.Terminal(
			os.Stdout,
			entries,
		)

	case "jsonl":
		return output.Pipeline(
			os.Stdout,
			entries,
		)

	default:
		return fmt.Errorf(
			"unknown format %q: expected tree or jsonl",
			format,
		)
	}
}
