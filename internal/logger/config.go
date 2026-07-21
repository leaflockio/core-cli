// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

// Format represents a log output format.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Level represents a log severity level.
type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarn    Level = "warn"
	LevelWarning Level = "warning" // Alias for LevelWarn.
	LevelError   Level = "error"
)

// Config holds all logger settings.
type Config struct {
	Level   Level         `mapstructure:"level"`
	Console ConsoleConfig `mapstructure:"console"`
	File    FileConfig    `mapstructure:"file"`
}

// ConsoleConfig configures console (stderr) logging output.
type ConsoleConfig struct {
	// Enabled controls whether logs are written to stderr.
	Enabled bool `mapstructure:"enabled"`

	// Format sets the output format: FormatText or FormatJSON.
	Format Format `mapstructure:"format"`
}

// FileConfig configures file-based logging with automatic rotation.
type FileConfig struct {
	// Enabled controls whether logs are written to a file.
	Enabled bool `mapstructure:"enabled"`

	// Format sets the output format: FormatText or FormatJSON.
	Format Format `mapstructure:"format"`

	// Path is the directory where log files are stored.
	Path string `mapstructure:"path"`

	// Filename is the name of the log file (e.g. "leaf.log").
	Filename string `mapstructure:"filename"`

	// MaxSizeMB is the maximum size in megabytes before the file is rotated.
	MaxSizeMB int `mapstructure:"max_size_mb"`

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `mapstructure:"max_backups"`

	// MaxAgeDays is the maximum number of days to retain old log files.
	// 0 means no age-based deletion.
	MaxAgeDays int `mapstructure:"max_age_days"`

	// Compress controls whether rotated log files are gzip-compressed.
	Compress bool `mapstructure:"compress"`
}
