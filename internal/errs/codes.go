// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package errs

// Error code convention:
//
// Format: <DOMAIN><NNN> — uppercase domain prefix, zero-padded three-digit number.
// Add new codes at the end of the relevant block, incrementing the number.
// Never reuse or renumber a code — retired codes stay as a comment.

// ENV domain covers errors caused by incorrect or missing environment
// configuration that the caller is responsible for providing. Use an ENV code
// when the problem is in how the environment was set up, not in leaf itself.
const (
	// ENV001 is used when a string is parsed as a runtime environment but
	// does not match any of the known values.
	ENV001 Code = "ENV001"
)

// INT domain covers unexpected internal failures that indicate a bug in leaf.
// These codes are reserved — only errs.Internal() assigns them. Callers must
// never construct an INT error directly.
const (
	// INT000 is the default code for an unclassified internal failure.
	INT000 Code = "INT000"
)
