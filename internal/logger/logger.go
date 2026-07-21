// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"io"
	"log/slog"
	"os"
)

// Build constructs a *slog.Logger from cfg.
// Console output is written to os.Stderr.
func Build(cfg *Config) (*slog.Logger, error) {
	return build(cfg, os.Stderr)
}

func build(cfg *Config, console io.Writer) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{Level: level}
	var handlers []slog.Handler

	if cfg.Console.Enabled {
		h, err := newHandler(cfg.Console.Format, console, opts)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, h)
	}

	if cfg.File.Enabled {
		fw, err := newFileWriter(&cfg.File)
		if err != nil {
			return nil, err
		}
		h, err := newHandler(cfg.File.Format, fw, opts)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, h)
	}

	if len(handlers) == 0 {
		return nil, ErrNoOutputs
	}

	if len(handlers) == 1 {
		return slog.New(handlers[0]), nil
	}
	return slog.New(&multiHandler{handlers: handlers}), nil
}
