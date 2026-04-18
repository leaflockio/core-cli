// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// errHandlerFailed is returned by failHandler to simulate a handler write failure.
var errHandlerFailed = errors.New("handler error")

// failHandler is a slog.Handler that always returns an error from Handle.
type failHandler struct{}

func (failHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

// slog.Handler interface mandates slog.Record by value — cannot use a pointer.
//
//nolint:gocritic
func (failHandler) Handle(_ context.Context, _ slog.Record) error { return errHandlerFailed }
func (failHandler) WithAttrs(_ []slog.Attr) slog.Handler          { return failHandler{} }
func (failHandler) WithGroup(_ string) slog.Handler               { return failHandler{} }

func TestNewHandler_text(t *testing.T) {
	h, err := newHandler(FormatText, io.Discard, nil)
	if err != nil {
		t.Fatalf("newHandler(text) error = %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewHandler_json(t *testing.T) {
	h, err := newHandler(FormatJSON, io.Discard, nil)
	if err != nil {
		t.Fatalf("newHandler(json) error = %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestNewHandler_invalidFormat(t *testing.T) {
	_, err := newHandler("xml", io.Discard, nil) // intentional invalid value
	if !errors.Is(err, ErrUnknownFormat) {
		t.Errorf("expected ErrUnknownFormat, got %v", err)
	}
}

func TestMultiHandler_Enabled_allDisabled(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelError}
	mh := &multiHandler{handlers: []slog.Handler{
		slog.NewTextHandler(io.Discard, opts),
		slog.NewTextHandler(io.Discard, opts),
	}}
	if mh.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expected Enabled to return false when all handlers reject the level")
	}
}

func TestMultiHandler_Handle_handlerError(t *testing.T) {
	mh := &multiHandler{handlers: []slog.Handler{
		failHandler{},
		slog.NewTextHandler(io.Discard, nil),
	}}
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	if err := mh.Handle(context.Background(), record); err == nil {
		t.Error("expected error from Handle when a sub-handler fails")
	}
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	log, err := build(bothOutputs(t.TempDir()), &buf)
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	log.With("key", "value").Info("attrs-test")
	if !strings.Contains(buf.String(), "attrs-test") {
		t.Errorf("expected console output to contain %q, got %q", "attrs-test", buf.String())
	}
}

func TestMultiHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	log, err := build(bothOutputs(t.TempDir()), &buf)
	if err != nil {
		t.Fatalf("build() error = %v", err)
	}
	log.WithGroup("grp").Info("group-test")
	if !strings.Contains(buf.String(), "group-test") {
		t.Errorf("expected console output to contain %q, got %q", "group-test", buf.String())
	}
}
