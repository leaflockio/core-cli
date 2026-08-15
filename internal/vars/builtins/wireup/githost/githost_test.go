// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package githost

import (
	"errors"
	"regexp"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/repo"
	"github.com/leaflockio/core-cli/internal/vars"
)

// errRemote stands in for a real remote-detection failure (err113
// disallows errors.New at the point of use).
var errRemote = errors.New("remote lookup failed")

func TestCompute_returnsHostWhenSet(t *testing.T) {
	a := &app.App{Repo: &repo.Info{Host: "github.com"}}

	got, err := compute(a)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if got != "github.com" {
		t.Errorf("compute = %q, want %q", got, "github.com")
	}
}

func TestCompute_propagatesRemoteErr(t *testing.T) {
	a := &app.App{Repo: &repo.Info{RemoteErr: errRemote}}

	_, err := compute(a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errRemote) {
		t.Errorf("error = %v, want wrapping %v", err, errRemote)
	}
}

func TestCompute_callerErrorWhenNoRemoteConfigured(t *testing.T) {
	a := &app.App{Repo: &repo.Info{}}

	_, err := compute(a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR004 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR004)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

func TestGitHost_volatilityIsStable(t *testing.T) {
	if GitHost.Volatility() != vars.Stable {
		t.Errorf("Volatility = %v, want Stable", GitHost.Volatility())
	}
}

func TestGitHost_patternMatchesRealisticHosts(t *testing.T) {
	re := regexp.MustCompile("^" + GitHost.Pattern() + "$")
	for _, tt := range []struct {
		in   string
		want bool
	}{
		{"github.com", true},
		{"gitlab.example.com", true},
		{"git-server", true},
		{"", false},
		{"has space", false},
		{"has/slash", false},
	} {
		if got := re.MatchString(tt.in); got != tt.want {
			t.Errorf("pattern.MatchString(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestGitHost_registersAndResolves(t *testing.T) {
	if GitHost == nil {
		t.Fatal("GitHost is nil — registration failed")
	}

	v, err := vars.New(&app.App{Repo: &repo.Info{Host: "github.com"}})
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	got, err := v.Resolve("{GIT_HOST}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "github.com" {
		t.Errorf("Resolve = %q, want %q", got, "github.com")
	}
}
