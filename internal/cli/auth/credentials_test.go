// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth_test

import (
	"testing"
	"time"

	"github.com/leaflock/core-cli/internal/cli/auth"
)

// TestIsExpired_zero verifies that credentials with a zero ExpiresAt never expire.
func TestIsExpired_zero(t *testing.T) {
	c := &auth.Credentials{Token: "tok"}
	if c.IsExpired() {
		t.Error("zero ExpiresAt should never be expired")
	}
}

// TestIsExpired_future verifies that credentials with a future ExpiresAt are not expired.
func TestIsExpired_future(t *testing.T) {
	c := &auth.Credentials{Token: "tok", ExpiresAt: time.Now().UTC().Add(time.Hour)}
	if c.IsExpired() {
		t.Error("future ExpiresAt should not be expired")
	}
}

// TestIsExpired_past verifies that credentials with a past ExpiresAt are expired.
func TestIsExpired_past(t *testing.T) {
	c := &auth.Credentials{Token: "tok", ExpiresAt: time.Now().UTC().Add(-time.Hour)}
	if !c.IsExpired() {
		t.Error("past ExpiresAt should be expired")
	}
}
