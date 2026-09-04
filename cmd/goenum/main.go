// Command goenum generates type-safe enum boilerplate for Go enums declared
// with //goenum directives.
//
// Usage:
//
//	goenum generate [path...]
//
// Paths may be Go source files or directories (directories are scanned
// non-recursively). With no paths, the current directory is scanned, which is
// what a //go:generate goenum directive needs.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/abdorrahmani/goenum/internal/generator"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "goenum:", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("goenum", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "usage: goenum generate [--force] [--dry-run] [path...]")
		fs.PrintDefaults()
	}
	force := fs.Bool("force", false, "overwrite a non-generated file at the output path")
	dryRun := fs.Bool("dry-run", false, "report changes without writing files")
	verbose := fs.Bool("v", false, "list every output file, including unchanged ones")

	fs.Parse(os.Args[1:])
	args := fs.Args()
	if len(args) == 0 || args[0] != "generate" {
		fs.Usage()
		return fmt.Errorf("first argument must be 'generate'")
	}
	// Flags may follow the subcommand: `goenum generate --force ./x`.
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	results, err := generator.Generate(generator.Options{
		Paths:  fs.Args(),
		Force:  *force,
		DryRun: *dryRun,
	})
	if err != nil {
		return err
	}
	written := 0
	for _, r := range results {
		if r.Changed {
			written++
			fmt.Fprintf(os.Stderr, "goenum: wrote %s\n", r.Path)
		} else if *verbose {
			fmt.Fprintf(os.Stderr, "goenum: up to date %s\n", r.Path)
		}
	}
	if written == 0 && !*verbose {
		fmt.Fprintln(os.Stderr, "goenum: no changes")
	}
	return nil
}
