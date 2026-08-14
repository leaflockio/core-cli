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

// FLR domain covers errors caused by invalid combinations of a command's
// flags — not specific to any one flag or command.
const (
	// FLR001 is used when two or more mutually exclusive flags are set at
	// the same time.
	FLR001 Code = "FLR001"

	// FLR002 is used when a flag is set without another flag it depends on
	// also being set.
	FLR002 Code = "FLR002"
)

// SRC domain covers errors resolving which files a command should operate
// on.
const (
	// SRC001 is used when positional file arguments are combined with
	// --staged or --pr — each already defines a complete file source on its
	// own, so combining them is ambiguous.
	SRC001 Code = "SRC001"

	// SRC002 is used when no source was given at all: no --staged, --pr, or
	// positional file arguments.
	SRC002 Code = "SRC002"
)

// GIT domain covers errors from the internal/git package — invalid or
// unavailable git state that an operation depends on, independent of which
// higher-level command triggered it.
const (
	// GIT001 is used when an operation requires being inside a git
	// repository, but the working directory isn't one.
	GIT001 Code = "GIT001"

	// GIT002 is used when a base ref is known but not resolvable in the
	// local git history.
	GIT002 Code = "GIT002"

	// GIT003 is used when there's no base ref to diff against at all — none
	// was given, and no remote/branch to build one from either.
	GIT003 Code = "GIT003"

	// GIT004 is used when a diff-filter string contains a character git's
	// own --diff-filter doesn't recognize.
	GIT004 Code = "GIT004"

	// GIT005 is used when a base ref begins with "-", which would let git
	// parse it as an option rather than a revision.
	GIT005 Code = "GIT005"

	// GIT006 is used when a file has no commit history to derive a date
	// from.
	GIT006 Code = "GIT006"
)

// WSP domain covers errors from the workspace package.
const (
	// WSP001 is used when the home directory cannot be resolved at workspace
	// construction time. This happens when os.UserHomeDir fails and no override
	// was provided.
	WSP001 Code = "WSP001"

	// WSP002 is used when the platform cache directory cannot be resolved.
	// Unlike WSP001, this does not fail workspace construction — it is
	// returned later, by Dir and File on the CommandSpace that ForCache
	// returns.
	WSP002 Code = "WSP002"
)

// CCF domain covers errors caused by how a directory of per-command config
// files is laid out: a flat manifest, or one file per command.
const (
	// CCF001 is used when the flat manifest file exists under more than one
	// discoverable extension — ambiguous, since there's no way to know
	// which one is meant to apply.
	CCF001 Code = "CCF001"

	// CCF002 is used when a flat manifest file coexists with one or more
	// per-command config files — ambiguous which layout is active.
	CCF002 Code = "CCF002"

	// CCF003 is used when the config directory exists but cannot be read
	// (e.g. a permissions error). A directory that does not exist at all is
	// not an error.
	CCF003 Code = "CCF003"

	// CCF004 is used when the command actually being run has its own config
	// file under more than one discoverable extension.
	CCF004 Code = "CCF004"

	// CCF005 is used when the command actually being run has a config file
	// that exists but could not be read or decoded.
	CCF005 Code = "CCF005"
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

// CLI domain covers errors caused by how the tool itself was invoked —
// independent of any specific command's own logic. Use a CLI code when the
// problem is in the command line the caller typed, not in a command's
// business rules.
const (
	// CLI001 is used when a command path segment doesn't match any known
	// subcommand.
	CLI001 Code = "CLI001"
)

// INT domain covers unexpected internal failures that indicate a bug in leaf.
// These codes are reserved — only errs.Internal() assigns them. Callers must
// never construct an INT error directly.
const (
	// INT000 is the default code for an unclassified internal failure.
	INT000 Code = "INT000"
)
