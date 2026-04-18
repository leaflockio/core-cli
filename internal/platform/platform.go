// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

// OS represents the host operating system.
type OS string

const (
	// MacOS is the macOS operating system.
	MacOS OS = "macos"
	// Linux is the Linux operating system.
	Linux OS = "linux"
	// Windows is the Windows operating system.
	Windows OS = "windows"
	// UnknownOS is used when the operating system cannot be determined.
	UnknownOS OS = "unknown"
)

// String returns the string representation of the OS.
func (o OS) String() string { return string(o) }

// Arch represents the host CPU architecture.
type Arch string

const (
	// AMD64 is the 64-bit x86 architecture.
	AMD64 Arch = "amd64"
	// ARM64 is the 64-bit ARM architecture.
	ARM64 Arch = "arm64"
	// UnknownArch is used when the architecture cannot be determined.
	UnknownArch Arch = "unknown"
)

// String returns the string representation of the Arch.
func (a Arch) String() string { return string(a) }

// PackageManager represents the system package manager available on the host.
type PackageManager string

const (
	// Homebrew is the macOS package manager.
	Homebrew PackageManager = "homebrew"
	// Apt is the Debian/Ubuntu package manager.
	Apt PackageManager = "apt"
	// UnknownPM is used when no supported package manager is found.
	UnknownPM PackageManager = "unknown"
)

// String returns the string representation of the PackageManager.
func (pm PackageManager) String() string { return string(pm) }

// Platform holds detected information about the host system.
type Platform struct {
	OS             OS
	Arch           Arch
	PackageManager PackageManager
	IsCI           bool
}
