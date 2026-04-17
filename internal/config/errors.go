// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import "errors"

// ErrUnknownEnv is returned when an unrecognized environment string is parsed.
var ErrUnknownEnv = errors.New("unknown env")

// ErrorsConfig holds error display settings.
type ErrorsConfig struct {
	ShowCode       bool `mapstructure:"show_code"`
	ShowResolution bool `mapstructure:"show_resolution"`
	ShowUnderlying bool `mapstructure:"show_underlying"`
}

const (
	keyErrorsShowCode       = "errors.show_code"
	keyErrorsShowResolution = "errors.show_resolution"
	keyErrorsShowUnderlying = "errors.show_underlying"
)
