// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package errs

// Code identifies an error by domain and sequence number.
type Code string

// Exit code constants.
const (
	ExitSuccess  = 0 // clean exit
	ExitUser     = 1 // caller error — bad input, missing configuration, missing tool
	ExitInternal = 2 // unexpected failure — indicates a bug
)

// Context pairs a cause with its resolution. Use multiple contexts when an
// error can stem from more than one root cause, each with a different fix.
type Context struct {
	Cause      string // what led to this error
	Resolution string // how to fix it
}

// Error is the central error type. It carries a stable code, a user-facing
// message, optional diagnostic contexts, an OS exit code, and the underlying
// Go error for errors.Is and errors.As chaining.
type Error struct {
	Code     Code
	Message  string
	Contexts []Context
	ExitCode int
	Err      error
}

// Error implements the error interface. Returns the human-readable message.
func (e *Error) Error() string { return e.Message }

// Unwrap returns the underlying error so errors.Is and errors.As work through
// the chain.
func (e *Error) Unwrap() error { return e.Err }

// Caller returns an Error for failures that are the caller's fault (exit 1).
// Use this when the caller provided bad input, a required tool is missing,
// or a config file is malformed.
func Caller(code Code, message string, err error, contexts ...Context) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Contexts: contexts,
		ExitCode: ExitUser,
		Err:      err,
	}
}

// Unexpected returns an Error for failures that should not have happened
// (exit 2). The message shown to the caller is generic; the underlying err
// is preserved for logging and support.
func Unexpected(err error, contexts ...Context) *Error {
	return &Error{
		Code:     INT000,
		Message:  "an unexpected error occurred",
		Contexts: contexts,
		ExitCode: ExitInternal,
		Err:      err,
	}
}
