// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import (
	"io"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

func newFileWriter(cfg *FileConfig) (io.Writer, error) {
	if cfg.Path == "" {
		return nil, ErrFilePathEmpty
	}
	return &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Path, cfg.Filename),
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}, nil
}
