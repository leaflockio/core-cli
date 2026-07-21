// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"strings"
	"testing"
)

func TestEnvPrefixMatchesAppName(t *testing.T) {
	if !strings.EqualFold(envPrefix, AppName) {
		t.Errorf("envPrefix %q does not match strings.ToUpper(AppName) %q — update both together in app.go",
			envPrefix, strings.ToUpper(AppName))
	}
}
