// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"testing"
)

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

// --- DetectProvider ---

func TestDetectProvider(t *testing.T) {
	// Set a CI env var to ensure we detect a provider
	t.Setenv(envGitHubActions, "true")
	p := DetectProvider()
	if p != ProviderGitHub {
		t.Errorf("expected ProviderGitHub, got %s", p)
	}
}
