// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package cli

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/spf13/cobra"
)

// ShowHelp resolves the command path in args from cmd's root and renders
// its help. An unrecognized topic is rejected through that command's own
// Args validator — the same one that rejects an unrecognized subcommand
// name during normal execution — instead of silently falling back to help
// for whichever command Find stopped at.
func ShowHelp(cmd *cobra.Command, args []string) error {
	found, leftover, err := cmd.Root().Find(args)
	if err != nil {
		return err
	}
	if err := found.ValidateArgs(leftover); err != nil {
		return err
	}
	return found.Help()
}

// unknownCommand rejects any positional arg that isn't a recognized
// subcommand of cmd, suggesting close matches when cobra can find any.
// It is Meta's default Args validator — a command that accepts its own
// positional arguments opts out via WithArgs.
func unknownCommand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}

	resolution := fmt.Sprintf("run %q to see the available commands", cmd.CommandPath()+" --help")
	if suggestions := suggestionsFor(cmd, args[0]); len(suggestions) > 0 {
		resolution = suggestionMessage(cmd, suggestions)
	}

	return errs.Caller(errs.CLI001,
		fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath()),
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("%q is not a subcommand of %q", args[0], cmd.CommandPath()),
			Resolution: resolution,
		},
	)
}

// suggestionMessage names the closest matching command(s) and the exact
// command to run, so the resolution is something the caller can act on
// directly rather than just a bare list of names.
func suggestionMessage(cmd *cobra.Command, suggestions []string) string {
	if len(suggestions) == 1 {
		return fmt.Sprintf("did you mean %q? run %q instead", suggestions[0], cmd.CommandPath()+" "+suggestions[0])
	}
	return fmt.Sprintf("did you mean one of: %s? run %q with the correct name",
		strings.Join(suggestions, ", "), cmd.CommandPath())
}

// suggestionsFor wraps cmd.SuggestionsFor, applying the same defaulting
// SuggestionsFor itself doesn't: a zero SuggestionsMinimumDistance would
// otherwise require a near-exact match, and DisableSuggestions must still be
// honored when a command opts out.
func suggestionsFor(cmd *cobra.Command, typedName string) []string {
	if cmd.DisableSuggestions {
		return nil
	}
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}
	return cmd.SuggestionsFor(typedName)
}
