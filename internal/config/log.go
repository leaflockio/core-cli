// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

// LogConfig holds logger settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

const (
	// KeyLogLevel is the viper key for log level.
	keyLogLevel = "log.level"

	// KeyLogFormat is the viper key for log format.
	keyLogFormat = "log.format"

	// DefaultLogFormat is the log format when no config file sets it.
	defaultLogFormat = "text"

	// DefaultLogLevelDev is the log level for dev and test profiles.
	defaultLogLevelDev = "debug"

	// DefaultLogLevelProd is the log level for the prod profile.
	defaultLogLevelProd = "info"
)
