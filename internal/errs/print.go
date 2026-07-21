// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package errs

import (
	"errors"
	"fmt"
	"os"
)

// DisplayConfig controls how errors are rendered to the user.
// All fields default to safe values so Print works before Configure is called.
type DisplayConfig struct {
	// ShowCode includes the error code (e.g. CFG001) in the output.
	// Useful for support and bug reports.
	ShowCode bool `mapstructure:"show_code"`

	// ShowResolution includes the resolution hint for each context.
	ShowResolution bool `mapstructure:"show_resolution"`

	// ShowUnderlying includes the raw wrapped Go error.
	// Intended for dev/debug use — raw errors are noise to end users.
	ShowUnderlying bool `mapstructure:"show_underlying"`
}

// defaults are the built-in fallback settings used before Configure is called.
// Chosen to be maximally helpful during early boot (before config loads).
var defaults = DisplayConfig{
	ShowCode:       true,
	ShowResolution: true,
	ShowUnderlying: false,
}

var active = defaults

// Configure updates the display settings from the loaded application config.
// Must be called after config.Load() succeeds. Safe to skip — defaults apply.
func Configure(cfg DisplayConfig) {
	active = cfg
}

// Print writes a structured, user-facing representation of err to stderr.
// If err is not an *Error, the message is printed as-is.
// Display behavior is controlled by Configure; defaults apply before it is called.
func Print(err error) {
	if err == nil {
		return
	}

	var e *Error
	if !errors.As(err, &e) {
		fmt.Fprintln(os.Stderr, "error:", err.Error())
		return
	}

	if active.ShowCode {
		fmt.Fprintf(os.Stderr, "error [%s]: %s\n", e.Code, e.Message)
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", e.Message)
	}

	for _, ctx := range e.Contexts {
		if ctx.Cause != "" {
			fmt.Fprintf(os.Stderr, "  cause: %s\n", ctx.Cause)
		}
		if active.ShowResolution && ctx.Resolution != "" {
			fmt.Fprintf(os.Stderr, "  resolution: %s\n", ctx.Resolution)
		}
	}

	if active.ShowUnderlying && e.Err != nil {
		fmt.Fprintf(os.Stderr, "  underlying: %s\n", e.Err.Error())
	}
}
