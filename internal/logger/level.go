// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

func parseLevel(l Level) (slog.Level, error) {
	switch Level(strings.ToLower(string(l))) {
	case LevelDebug:
		return slog.LevelDebug, nil
	case LevelInfo:
		return slog.LevelInfo, nil
	case LevelWarn, LevelWarning:
		return slog.LevelWarn, nil
	case LevelError:
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("%w: %s", ErrUnknownLevel, l)
	}
}
