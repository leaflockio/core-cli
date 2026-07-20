// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/logger"
	"github.com/spf13/viper"
)

const (
	// ExtHintOpen, extHintSep, extHintClose are used to format multi-extension
	// hints in error messages, e.g. {yaml,yml}.
	extHintOpen  = "{"
	extHintSep   = ","
	extHintClose = "}"

	// ConfigKeySep is the separator used in config key names (e.g. log.level).
	configKeySep = "."

	// EnvKeySep is the separator used in environment variable names (e.g. LEAF_LOG_LEVEL).
	envKeySep = "_"
)

// Config holds all application configuration.
type Config struct {
	Log    logger.Config `mapstructure:"log"`
	Errors ErrorsConfig  `mapstructure:"errors"`
}

// Load reads configuration for env.
//
// Loading order (each layer overrides the previous):
//  1. Code defaults
//  2. config.yaml (base file)
//  3. config.{env}.yaml (env overlay)
//  4. LEAF_* environment variables
func Load(env Env) (*Config, error) {
	v := initViper(env)

	dirs := configDirs(env)
	exts := configFileExtensions(configFileType)
	debugf(levelDebug, "search paths: %v", dirs)

	if err := loadBase(v, dirs, exts); err != nil {
		return nil, err
	}
	if err := loadOverlay(v, env, dirs, exts); err != nil {
		return nil, err
	}
	return unmarshalConfig(v)
}

func initViper(env Env) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(configKeySep, envKeySep))
	v.AutomaticEnv()
	v.SetConfigType(configFileType)
	setDefaults(v, env)
	return v
}

func loadBase(v *viper.Viper, dirs, exts []string) error {
	path, ok := findConfigFile(dirs, configFileName, exts)
	if !ok {
		debugf(levelWarn, "no %s.%s found, using defaults", configFileName, formatExtHint(exts))
		return nil
	}
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		fmtType := strings.ToUpper(configFileType)
		cause := fmt.Sprintf("the base config file could not be read or contains invalid %s", fmtType)
		return errs.Caller(
			errs.CFG001,
			"failed to read base config file",
			err,
			errs.Context{
				Cause:      cause,
				Resolution: fmt.Sprintf("check %s.%s for syntax errors", configFileName, formatExtHint(exts)),
			},
		)
	}
	debugf(levelDebug, "loaded base config: %s", v.ConfigFileUsed())
	return nil
}

func loadOverlay(v *viper.Viper, env Env, dirs, exts []string) error {
	overlayName := fmt.Sprintf("%s.%s", configFileName, env)
	path, ok := findConfigFile(dirs, overlayName, exts)
	if !ok {
		debugf(levelDebug, "no %s overlay found", overlayName)
		return nil
	}
	v.SetConfigFile(path)
	if err := v.MergeInConfig(); err != nil {
		fmtType := strings.ToUpper(configFileType)
		cause := fmt.Sprintf("the %s overlay config file could not be read or contains invalid %s", env, fmtType)
		return errs.Caller(
			errs.CFG002,
			fmt.Sprintf("failed to read %s overlay config file", env),
			err,
			errs.Context{
				Cause:      cause,
				Resolution: fmt.Sprintf("check %s.%s.%s for syntax errors", configFileName, env, formatExtHint(exts)),
			},
		)
	}
	debugf(levelDebug, "merged %s overlay: %s", env, v.ConfigFileUsed())
	return nil
}

func unmarshalConfig(v *viper.Viper) (*Config, error) {
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errs.Caller(
			errs.CFG003,
			"config structure does not match the expected schema",
			err,
			errs.Context{
				Cause:      "your config file may contain invalid keys or wrong value types",
				Resolution: "check the docs for the correct config schema",
			},
			errs.Context{
				Cause:      "if you recently updated leaf, the config schema may have changed",
				Resolution: "check the changelog for schema changes and migrate your config",
			},
		)
	}
	return &cfg, nil
}

// formatExtHint formats a list of extensions for use in error messages.
// Multiple extensions are wrapped in braces: {yaml,yml}. A single extension
// is returned as-is: json.
func formatExtHint(exts []string) string {
	if len(exts) == 1 {
		return exts[0]
	}
	return extHintOpen + strings.Join(exts, extHintSep) + extHintClose
}

func setDefaults(v *viper.Viper, env Env) {
	switch env {
	case EnvDev, EnvTest:
		v.SetDefault(keyLogLevel, defaultLogLevelDev)
		v.SetDefault(keyLogConsoleEnabled, true)
		v.SetDefault(keyLogConsoleFormat, logger.FormatText)
		v.SetDefault(keyLogFileEnabled, false)
		v.SetDefault(keyLogFileFormat, logger.FormatText)
	case EnvProd:
		v.SetDefault(keyLogLevel, defaultLogLevelProd)
		v.SetDefault(keyLogConsoleEnabled, false)
		v.SetDefault(keyLogConsoleFormat, logger.FormatText)
		v.SetDefault(keyLogFileEnabled, true)
		v.SetDefault(keyLogFileFormat, logger.FormatJSON)
		v.SetDefault(keyLogFilePath, defaultLogPath())
	}

	v.SetDefault(keyLogFileFilename, defaultLogFilename)
	v.SetDefault(keyLogFileMaxSizeMB, defaultLogMaxSizeMB)
	v.SetDefault(keyLogFileMaxBackups, defaultLogMaxBackups)
	v.SetDefault(keyLogFileMaxAgeDays, defaultLogMaxAgeDays)
	v.SetDefault(keyLogFileCompress, true)

	v.SetDefault(keyErrorsShowCode, true)
	v.SetDefault(keyErrorsShowResolution, true)
	v.SetDefault(keyErrorsShowUnderlying, false)
}
