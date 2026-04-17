// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

const (
	// AppName is the CLI binary name. This is the single source of truth —
	// update this if the binary is ever renamed.
	appName = "leaf"

	// AppFolder is the hidden directory name used in home and XDG paths.
	// Derived from appName with a leading dot.
	appFolder = "." + appName

	// EnvPrefix is the prefix for all LEAF_* environment variables.
	// Must be the uppercase form of appName — update together.
	envPrefix = "LEAF"

	// EnvVarEnv selects the active profile (dev, test, prod).
	envVarEnv = envPrefix + "_ENV"

	// EnvVarConfigDir overrides the config search directory in dev/test.
	envVarConfigDir = envPrefix + "_CONFIG_DIR"

	// EnvVarConfigDebug enables debug output from the config loader.
	envVarConfigDebug = envPrefix + "_CONFIG_DEBUG"

	// ConfigFileName is the base name for config files:
	// config.yaml, config.test.yaml, etc.
	configFileName = "config"

	// ConfigDirName is the config subdirectory name within a project.
	configDirName = "config"

	// ConfigFileType is the file format viper expects.
	configFileType = "yaml"

	// YamlExt and yamlExtShort are the two equivalent file extensions for YAML.
	// Both are accepted regardless of which one configFileType is set to.
	yamlExt      = "yaml"
	yamlExtShort = "yml"
)
