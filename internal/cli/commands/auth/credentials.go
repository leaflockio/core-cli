// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package auth

import "time"

// Credentials holds everything the CLI needs to authenticate with the API.
type Credentials struct {
	Token     string    `mapstructure:"token"`
	ExpiresAt time.Time `mapstructure:"expires_at"`
	Workspace string    `mapstructure:"workspace"`
}

// IsExpired reports whether the credentials have passed their expiry time.
// This is a local pre-check only — the API is the authority and will return
// 401 regardless of what this returns. Credentials with a zero ExpiresAt
// never expire (PATs).
func (c *Credentials) IsExpired() bool {
	return !c.ExpiresAt.IsZero() && time.Now().UTC().After(c.ExpiresAt.UTC())
}
