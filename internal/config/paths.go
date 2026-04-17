// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"os"
	"path/filepath"

	"github.com/leaflock/core-cli/internal/util/pathutil"
)

const (
	// Standard XDG env var for the user config directory.
	xdgConfigHomeEnvVar = "XDG_CONFIG_HOME"

	// Default XDG config base directory when XDG_CONFIG_HOME is unset.
	xdgDefaultConfigDir = ".config"
)

// osExecutable and osUserHomeDir are extracted as variables so tests can
// inject errors without depending on real OS state.
var (
	osExecutable  = os.Executable
	osUserHomeDir = os.UserHomeDir
)

// executableDir returns the directory of the running executable.
// Extracted as a variable so tests can override it without depending on
// the real binary path.
var executableDir = func() string {
	exe, err := osExecutable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

func configDirs(env Env) []string {
	if env == EnvProd {
		return xdgConfigDirs()
	}
	return devConfigDirs()
}

func xdgConfigDirs() []string {
	base := os.Getenv(xdgConfigHomeEnvVar)
	if base == "" {
		home, err := osUserHomeDir()
		if err != nil {
			return []string{}
		}
		base = filepath.Join(home, xdgDefaultConfigDir)
	}
	return []string{filepath.Join(base, appFolder)}
}

func devConfigDirs() []string {
	if d := os.Getenv(envVarConfigDir); d != "" {
		debugf(levelDebug, "using %s override: %s", envVarConfigDir, d)
		return []string{d}
	}
	if root, ok := findProjectRoot(); ok {
		return []string{filepath.Join(root, configDirName)}
	}
	return []string{configDirName}
}

// configFileExtensions returns all valid file extensions for fileType.
// For yaml types, both "yaml" and "yml" are returned. All other types
// return a single-element slice with the type itself.
func configFileExtensions(fileType string) []string {
	switch fileType {
	case yamlExt, yamlExtShort:
		return []string{yamlExt, yamlExtShort}
	default:
		return []string{fileType}
	}
}

// findConfigFile searches dirs for the first file matching name.ext for each
// ext in exts. Returns the full path and true if found.
func findConfigFile(dirs []string, name string, exts []string) (string, bool) {
	for _, dir := range dirs {
		for _, ext := range exts {
			p := filepath.Join(dir, name+"."+ext)
			if _, err := os.Stat(p); err == nil {
				return p, true
			}
		}
	}
	return "", false
}

// findProjectRoot locates the project root by walking up from the executable.
func findProjectRoot() (string, bool) {
	return pathutil.FindModuleRoot(executableDir())
}
