// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"errors"
	"testing"
)

var errNotFound = errors.New("executable file not found in PATH")

func notFound(_ string) (string, error) { return "", errNotFound }
func found(_ string) (string, error)    { return "tool", nil }
func noEnv(_ string) string             { return "" }

// --- Detect ---

func TestDetect_returnsNonNil(t *testing.T) {
	if Detect() == nil {
		t.Fatal("expected non-nil Platform from Detect()")
	}
}

// --- resolveOS ---

func TestResolveOS_macOS(t *testing.T) {
	if resolveOS(goosDarwin) != MacOS {
		t.Errorf("expected MacOS, got %s", resolveOS(goosDarwin))
	}
}

func TestResolveOS_linux(t *testing.T) {
	if resolveOS(goosLinux) != Linux {
		t.Errorf("expected Linux, got %s", resolveOS(goosLinux))
	}
}

func TestResolveOS_windows(t *testing.T) {
	if resolveOS(goosWindows) != Windows {
		t.Errorf("expected Windows, got %s", resolveOS(goosWindows))
	}
}

func TestResolveOS_unknown(t *testing.T) {
	if resolveOS("plan9") != UnknownOS {
		t.Errorf("expected UnknownOS, got %s", resolveOS("plan9"))
	}
}

// --- resolveArch ---

func TestResolveArch_amd64(t *testing.T) {
	if resolveArch(goarchAMD64) != AMD64 {
		t.Errorf("expected AMD64, got %s", resolveArch(goarchAMD64))
	}
}

func TestResolveArch_arm64(t *testing.T) {
	if resolveArch(goarchARM64) != ARM64 {
		t.Errorf("expected ARM64, got %s", resolveArch(goarchARM64))
	}
}

func TestResolveArch_unknown(t *testing.T) {
	if resolveArch("mips") != UnknownArch {
		t.Errorf("expected UnknownArch, got %s", resolveArch("mips"))
	}
}

// --- resolvePackageManager ---

func TestResolvePackageManager_macOSWithBrew(t *testing.T) {
	if resolvePackageManager(MacOS, found) != Homebrew {
		t.Error("expected Homebrew when brew is on PATH")
	}
}

func TestResolvePackageManager_macOSWithoutBrew(t *testing.T) {
	if resolvePackageManager(MacOS, notFound) != UnknownPM {
		t.Error("expected UnknownPM when brew is not on PATH")
	}
}

func TestResolvePackageManager_linuxWithApt(t *testing.T) {
	if resolvePackageManager(Linux, found) != Apt {
		t.Error("expected Apt when apt is on PATH")
	}
}

func TestResolvePackageManager_linuxWithoutApt(t *testing.T) {
	if resolvePackageManager(Linux, notFound) != UnknownPM {
		t.Error("expected UnknownPM when apt is not on PATH")
	}
}

func TestResolvePackageManager_windows(t *testing.T) {
	if resolvePackageManager(Windows, notFound) != UnknownPM {
		t.Error("expected UnknownPM for Windows")
	}
}

func TestResolvePackageManager_unknown(t *testing.T) {
	if resolvePackageManager(UnknownOS, notFound) != UnknownPM {
		t.Error("expected UnknownPM for UnknownOS")
	}
}

// --- resolveIsCI ---

func TestResolveIsCI_githubActions(t *testing.T) {
	result := resolveIsCI(func(key string) string {
		if key == envGitHubActions {
			return envGitHubActionsVal
		}
		return ""
	})
	if !result {
		t.Error("expected IsCI=true when GITHUB_ACTIONS=true")
	}
}

func TestResolveIsCI_genericCI(t *testing.T) {
	result := resolveIsCI(func(key string) string {
		if key == envCI {
			return "true"
		}
		return ""
	})
	if !result {
		t.Error("expected IsCI=true when CI is set")
	}
}

func TestResolveIsCI_notCI(t *testing.T) {
	if resolveIsCI(noEnv) {
		t.Error("expected IsCI=false when no CI env vars are set")
	}
}
