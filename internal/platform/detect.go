// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"os"
	"os/exec"
	"runtime"
)

// runtime.GOOS values.
const (
	goosDarwin  = "darwin"
	goosLinux   = "linux"
	goosWindows = "windows"
)

// runtime.GOARCH values.
const (
	goarchAMD64 = "amd64"
	goarchARM64 = "arm64"
)

// Package manager binary names used for PATH detection.
const (
	binBrew = "brew"
	binApt  = "apt"
)

// CI environment variables.
const (
	envGitHubActions    = "GITHUB_ACTIONS"
	envGitHubActionsVal = "true"
	envCI               = "CI"
)

// Detect returns a Platform populated with information about the current host.
func Detect() *Platform {
	return detect(runtime.GOOS, runtime.GOARCH, exec.LookPath, os.Getenv)
}

func detect(
	goos, goarch string,
	lookPath func(string) (string, error),
	getenv func(string) string,
) *Platform {
	o := resolveOS(goos)

	return &Platform{
		OS:             o,
		Arch:           resolveArch(goarch),
		PackageManager: resolvePackageManager(o, lookPath),
		IsCI:           resolveIsCI(getenv),
	}
}

func resolveOS(goos string) OS {
	switch goos {
	case goosDarwin:
		return MacOS
	case goosLinux:
		return Linux
	case goosWindows:
		return Windows
	default:
		return UnknownOS
	}
}

func resolveArch(goarch string) Arch {
	switch goarch {
	case goarchAMD64:
		return AMD64
	case goarchARM64:
		return ARM64
	default:
		return UnknownArch
	}
}

func resolvePackageManager(o OS, lookPath func(string) (string, error)) PackageManager {
	switch o {
	case MacOS:
		if _, err := lookPath(binBrew); err == nil {
			return Homebrew
		}
	case Linux:
		if _, err := lookPath(binApt); err == nil {
			return Apt
		}
	case Windows, UnknownOS:
		// no supported package manager
	}

	return UnknownPM
}

// resolveIsCI reports whether the process is running inside a CI environment.
// It checks GITHUB_ACTIONS (GitHub Actions) and the generic CI variable set
// by most CI systems (CircleCI, GitLab CI, Travis, etc.).
func resolveIsCI(getenv func(string) string) bool {
	return getenv(envGitHubActions) == envGitHubActionsVal || getenv(envCI) != ""
}
