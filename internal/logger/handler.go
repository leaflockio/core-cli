// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

func newHandler(format Format, w io.Writer, opts *slog.HandlerOptions) (slog.Handler, error) {
	switch format {
	case FormatText:
		return slog.NewTextHandler(w, opts), nil
	case FormatJSON:
		return slog.NewJSONHandler(w, opts), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownFormat, format)
	}
}

// multiHandler fans out log records to multiple slog.Handlers, each with
// its own format. Used when both console and file outputs are enabled.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// slog.Handler interface mandates slog.Record by value — we cannot use a pointer
// here even though gocritic flags it as a heavy parameter (288 bytes).
//
//nolint:gocritic
func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
