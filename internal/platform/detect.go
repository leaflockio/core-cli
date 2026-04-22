// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"os"
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

// Detect returns a Platform populated with information about the current host.
func Detect() *Platform {
	o := resolveOS(runtime.GOOS)

	return &Platform{
		OS:              o,
		Arch:            resolveArch(runtime.GOARCH),
		PackageManagers: NewPkgDetector(o).Detect(),
		Container:       NewContainerDetector().Detect(),
		CI:              NewCIDetector(os.Getenv).Detect(),
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

// DetectProvider returns the active CI provider by inspecting environment
// variables.
func DetectProvider() Provider {
	return NewCIDetector(os.Getenv).Detect().Provider
}
