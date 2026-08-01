// Command gen-llmdocs regenerates the generated chapters of the embedded LLM
// reference. It is run by go generate, and re-run in CI to prove the committed
// artifact still matches the command tree.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ideamans/go-llm-cli-kit/catalog"

	"github.com/ideamans/mautic-cli/cmd"
)

const outPath = "90-commands.md"

func main() {
	// Root returns the fully assembled cobra tree without executing it.
	root := cmd.Root()

	md := catalog.Markdown(root, catalog.Options{
		Title: "Command catalog",
		Intro: "Generated from the cobra command tree by `go generate ./...`. " +
			"Do not edit by hand — edit the command definitions instead.\n\n" +
			"Global flags apply to every command and are listed once, under `mautic`.",
		// The llm subcommand documents itself in 00-guide.md; listing it here
		// only wastes context.
		Skip: []string{"llm"},
	})

	dir, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	// go generate runs in the directory of the file carrying the directive.
	out := filepath.Join(dir, outPath)
	if err := os.WriteFile(out, []byte(md), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", out, len(md))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gen-llmdocs:", err)
	os.Exit(1)
}
