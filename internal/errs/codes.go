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

// CFG domain covers errors caused by malformed or unreadable configuration
// files. Use a CFG code when the problem is in the config file content, not
// in leaf itself.
const (
	// CFG001 is used when the base config file exists but cannot be read or
	// contains invalid YAML.
	CFG001 Code = "CFG001"

	// CFG002 is used when an environment overlay config file exists but cannot
	// be read or contains invalid YAML.
	CFG002 Code = "CFG002"

	// CFG003 is used when the config file is valid but its structure does not
	// match the expected schema — either due to a user mistake or a schema
	// change after a product update.
	CFG003 Code = "CFG003"
)

// INT domain covers unexpected internal failures that indicate a bug in leaf.
// These codes are reserved — only errs.Internal() assigns them. Callers must
// never construct an INT error directly.
const (
	// INT000 is the default code for an unclassified internal failure.
	INT000 Code = "INT000"
)
