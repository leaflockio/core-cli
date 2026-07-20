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

// LIC domain covers errors from the leaf license command.
const (
	// LIC001 is used when license-template.lock cannot be read, parsed, or written.
	LIC001 Code = "LIC001"

	// LIC002 is used when the template cannot be loaded from the configured source
	// (built-in name not found, local file missing, remote fetch failed).
	LIC002 Code = "LIC002"

	// LIC003 is used when a {VARIABLE} placeholder in the template is not resolved
	// by any source (built-ins, config vars, env vars, or --var flags).
	LIC003 Code = "LIC003"

	// LIC004 is used when a source file cannot be read or written during
	// add, update, or migrate operations.
	LIC004 Code = "LIC004"

	// LIC005 is used when --staged or --pr is requested but the working directory
	// is not inside a git repository.
	LIC005 Code = "LIC005"

	// LIC006 is used when the --pr base ref cannot be resolved or fetched from
	// the remote. Includes instructions for the caller to fix the issue.
	LIC006 Code = "LIC006"

	// LIC007 is used when leaf.yaml exists but its license section cannot be
	// parsed or contains invalid values.
	LIC007 Code = "LIC007"

	// LIC008 is used when one or more files fail the license header check.
	// This is a validation error (exit 3), not a configuration error.
	LIC008 Code = "LIC008"
)

// SCO domain covers errors caused by invalid combinations or usage of the
// shared scope flags (--all, --staged, --pr, --base). Use an SCO code when
// the problem is in how the user combined scope flags, not in the operation
// itself.
const (
	// SCO001 is used when more than one of --all, --staged, and --pr are
	// provided at the same time. These flags are mutually exclusive.
	SCO001 Code = "SCO001"

	// SCO002 is used when --base is provided without --pr. The --base flag
	// only applies to --pr scope resolution.
	SCO002 Code = "SCO002"
)

// WSP domain covers errors from the workspace package.
const (
	// WSP001 is used when the home directory cannot be resolved at workspace
	// construction time. This happens when os.UserHomeDir fails and no override
	// was provided.
	WSP001 Code = "WSP001"
)

// AUT domain covers errors from the auth package.
const (
	// AUT001 is used when no credentials are found in the active backend.
	AUT001 Code = "AUT001"

	// AUT002 is used when credentials cannot be read from the active backend
	// due to an I/O error or keychain permission denial.
	AUT002 Code = "AUT002"

	// AUT003 is used when credentials cannot be written to the active backend
	// due to an I/O error or keychain permission denial.
	AUT003 Code = "AUT003"

	// AUT004 is used when credentials cannot be deleted from the active backend.
	AUT004 Code = "AUT004"

	// AUT005 is used when stored credentials have passed their expiry time.
	AUT005 Code = "AUT005"
)

// INT domain covers unexpected internal failures that indicate a bug in leaf.
// These codes are reserved — only errs.Internal() assigns them. Callers must
// never construct an INT error directly.
const (
	// INT000 is the default code for an unclassified internal failure.
	INT000 Code = "INT000"
)
