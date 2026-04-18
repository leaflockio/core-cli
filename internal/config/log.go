// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import "github.com/leaflock/core-cli/internal/logger"

const (
	keyLogLevel = "log.level"

	keyLogConsoleEnabled = "log.console.enabled"
	keyLogConsoleFormat  = "log.console.format"

	keyLogFileEnabled    = "log.file.enabled"
	keyLogFileFormat     = "log.file.format"
	keyLogFileFilename   = "log.file.filename"
	keyLogFileMaxSizeMB  = "log.file.max_size_mb"
	keyLogFileMaxBackups = "log.file.max_backups"
	keyLogFileMaxAgeDays = "log.file.max_age_days"
	keyLogFileCompress   = "log.file.compress"

	defaultLogLevelDev   = logger.LevelDebug
	defaultLogLevelProd  = logger.LevelInfo
	defaultLogFilename   = appName + ".log"
	defaultLogMaxSizeMB  = 100
	defaultLogMaxBackups = 3
	defaultLogMaxAgeDays = 28
)
