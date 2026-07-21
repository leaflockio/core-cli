// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package cli defines the framework types a command declares itself with:
// [Command], [Definition], [Meta], and [GroupID].
//
// # Definition is declarative, not direct cobra wiring
//
// [Command.Define] returns a complete [Definition] instead of building a
// cobra command directly. This lets the whole declaration be validated as a
// unit before anything is wired, such as checking for duplicate flag names
// or requiring a [Definition.Handler] unless the command declares children.
//
// # Group defaults and display order live together
//
// [GroupID]'s values and the [Groups] slice that orders them for display are
// declared side by side, so adding or reordering a group is a single edit.
// [NewDefinition] defaults every command to [GroupCLI], so only commands
// that need a different placement have to opt in.
//
// # Every command package exposes New(...) Command
//
// A package implementing [Command] exposes a constructor named New,
// returning [Command] rather than its own concrete type — even when it takes
// no arguments. This keeps every call site the same shape regardless of
// whether a given command needs construction-time dependencies, and keeps
// the concrete type unexported.
package cli
