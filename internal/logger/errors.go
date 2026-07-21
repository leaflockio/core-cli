// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package logger

import "errors"

// Sentinel errors for logger construction.
var (
	ErrNoOutputs     = errors.New("no log outputs enabled")
	ErrUnknownLevel  = errors.New("unknown log level")
	ErrUnknownFormat = errors.New("unknown log format")
	ErrFilePathEmpty = errors.New("log file path is empty")
)
