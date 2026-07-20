// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"os"
	"path/filepath"

	"github.com/leaflockio/core-cli/internal/logger"
)

const (
	keyLogLevel = "log.level"

	keyLogConsoleEnabled = "log.console.enabled"
	keyLogConsoleFormat  = "log.console.format"

	keyLogFileEnabled    = "log.file.enabled"
	keyLogFileFormat     = "log.file.format"
	keyLogFilePath       = "log.file.path"
	keyLogFileFilename   = "log.file.filename"
	keyLogFileMaxSizeMB  = "log.file.max_size_mb"
	keyLogFileMaxBackups = "log.file.max_backups"
	keyLogFileMaxAgeDays = "log.file.max_age_days"
	keyLogFileCompress   = "log.file.compress"

	// XdgStateHomeEnvVar is the XDG env var for state data (logs, history).
	xdgStateHomeEnvVar = "XDG_STATE_HOME"

	// XdgDefaultStateDir is the default XDG state directory relative to home.
	xdgDefaultStateDir = ".local/state"

	defaultLogLevelDev   = logger.LevelDebug
	defaultLogLevelProd  = logger.LevelInfo
	defaultLogFilename   = AppName + ".log"
	defaultLogMaxSizeMB  = 100
	defaultLogMaxBackups = 3
	defaultLogMaxAgeDays = 28
)

// defaultLogPath resolves the log directory via XDG_STATE_HOME, falling back
// to ~/.local/state if the env var is unset, per the XDG Base Dir spec.
func defaultLogPath() string {
	base := os.Getenv(xdgStateHomeEnvVar)
	if base == "" {
		home, err := osUserHomeDir()
		if err != nil {
			return "."
		}
		base = filepath.Join(home, xdgDefaultStateDir)
	}
	return filepath.Join(base, AppName)
}
