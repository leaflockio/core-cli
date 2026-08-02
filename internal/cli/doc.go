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
// # Every command package exposes New(...) Command
//
// A package implementing [Command] exposes a constructor named New,
// returning [Command] rather than its own concrete type — even when it takes
// no arguments. This keeps every call site the same shape regardless of
// whether a given command needs construction-time dependencies, and keeps
// the concrete type unexported.
package cli
