// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/leaflock/core-cli/internal/util/pathutil"
)

type level string

const (
	levelDebug level = "DEBUG"
	levelWarn  level = "WARN"
)

const (
	prefixOpen  = "["
	prefixClose = "]"
)

const (
	debugLevelOff  = 0 // no output
	debugLevelPkg  = 1 // package-level prefix
	debugLevelFile = 2 // file-level prefix
)

const debugLineFormat = "%s %-5s %s\n"

const debugWarnLevelFmt = "warning: %s must be %d (off), %d (package), or %d (file), got %q — debug disabled\n"

var (
	debugLevel     int
	debugPkgPrefix string
)

func init() {
	resetDebug(os.Getenv(envVarConfigDebug))
}

// resetDebug re-initializes debug state from levelStr. Extracted from init
// so tests can exercise both branches without restarting the process.
func resetDebug(levelStr string) {
	var err error
	debugLevel, err = strconv.Atoi(levelStr)
	if err != nil && levelStr != "" {
		fmt.Fprintf(os.Stderr, debugWarnLevelFmt,
			envVarConfigDebug, debugLevelOff, debugLevelPkg, debugLevelFile, levelStr)
		debugLevel = debugLevelOff
	}
	debugPkgPrefix = ""
	if debugLevel >= debugLevelPkg {
		_, file, _, _ := runtime.Caller(0)
		debugPkgPrefix = formatPrefix(path.Dir(pathutil.RelPath(file)))
	}
}

func debugf(lvl level, format string, args ...any) {
	if debugLevel == debugLevelOff {
		return
	}
	prefix := debugPkgPrefix
	if debugLevel >= debugLevelFile {
		_, file, _, _ := runtime.Caller(1)
		prefix = formatPrefix(pathutil.RelPath(file))
	}
	fmt.Fprintf(os.Stderr, debugLineFormat, prefix, string(lvl), fmt.Sprintf(format, args...))
}

func formatPrefix(p string) string {
	return prefixOpen + p + prefixClose
}
