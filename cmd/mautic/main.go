// Command mautic is a command line client for the Mautic REST API.
//
// The tree itself lives in the cmd package so that the catalog generator can
// import it and render internal/llmdocs/90-commands.md without running it.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/mautic-cli/cmd"
	"github.com/ideamans/mautic-cli/internal/llmdocs"
)

func main() {
	cfg := llmcmd.Config{Docs: llmdocs.Docs()}

	// The deprecated --llm flag has to work at any position on the command
	// line, so it is handled before cobra gets a chance to resolve
	// subcommands.
	if handled, err := llmcmd.HandleLegacy(os.Args[1:], cfg, os.Stdout); handled {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cmd.Root()
	llmcmd.AddTo(root, cfg)

	if err := root.ExecuteContext(ctx); err != nil {
		// Commands that have already printed a full explanation return a
		// silent error; anything else gets the standard prefix.
		if msg := err.Error(); msg != "" {
			fmt.Fprintln(os.Stderr, "Error:", msg)
		}
		if errors.Is(err, context.Canceled) {
			os.Exit(130)
		}
		os.Exit(1)
	}
}
