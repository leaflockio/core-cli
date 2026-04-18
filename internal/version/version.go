// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package version

// Build-time variables injected via -ldflags. Default values are used in
// development builds when no flags are provided.
//
//	-X github.com/leaflock/core-cli/internal/version.Version=1.0.0
//	-X github.com/leaflock/core-cli/internal/version.Commit=abc1234
//	-X github.com/leaflock/core-cli/internal/version.Date=2026-04-18
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info holds build-time metadata about the binary.
type Info struct {
	Version string // Semver string, e.g. "1.2.0".
	Commit  string // Git commit SHA at build time.
	Date    string // Build date.
}

// Current returns an Info populated from the build-time variables.
func Current() *Info {
	return &Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}
