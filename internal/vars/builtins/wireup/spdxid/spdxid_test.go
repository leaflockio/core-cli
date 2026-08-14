// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package spdxid

import (
	"errors"
	"testing"

	"github.com/leaflockio/core-cli/internal/app"
	"github.com/leaflockio/core-cli/internal/errs"
	"github.com/leaflockio/core-cli/internal/repo"
	"github.com/leaflockio/core-cli/internal/vars"
)

func TestCompute_returnsSPDXIDWhenSet(t *testing.T) {
	a := &app.App{Repo: &repo.Info{License: repo.LicenseInfo{SPDXID: "MIT"}}}

	got, err := compute(a)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if got != "MIT" {
		t.Errorf("compute = %q, want %q", got, "MIT")
	}
}

func TestCompute_callerErrorWhenNoLicenseClassified(t *testing.T) {
	a := &app.App{Repo: &repo.Info{}}

	_, err := compute(a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is not an *errs.Error: %v", err)
	}
	if e.Code != errs.VAR005 {
		t.Errorf("Code = %q, want %q", e.Code, errs.VAR005)
	}
	if e.ExitCode != errs.ExitUser {
		t.Errorf("ExitCode = %d, want %d (ExitUser)", e.ExitCode, errs.ExitUser)
	}
}

func TestSPDXID_registersAndResolves(t *testing.T) {
	if SPDXID == nil {
		t.Fatal("SPDXID is nil — registration failed")
	}

	v, err := vars.New(&app.App{Repo: &repo.Info{License: repo.LicenseInfo{SPDXID: "MIT"}}})
	if err != nil {
		t.Fatalf("vars.New: %v", err)
	}
	got, err := v.Resolve("{SPDX_ID}")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != "MIT" {
		t.Errorf("Resolve = %q, want %q", got, "MIT")
	}
}
