// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package factory builds a cobra command tree from cli.Command definitions.
//
// # Local flags vs. persistent flags
//
// assemble splits a Definition's flags into two plans: flags, built from the
// command's own Definition.Flags and registered on that command's local
// Flags(); and persistentFlags, built from the engine's implicitSystemFlags
// and registered only once, on the root's PersistentFlags, by Build. Cobra
// already propagates persistent flags to every descendant, so registering
// them on each command individually would just duplicate them — once under
// that command's own "Flags:" and again under its children's.
//
// # Build wires the whole tree in one call
//
// Build assembles and wires the command it's given along with every
// descendant, and is the only place that registers persistentFlags. Every
// command in the tree beyond the one passed to Build is reached through
// buildNode, an internal helper that both Build (once, for the top-level
// command) and execute (once per child, recursively) call — but only Build
// registers persistent flags, so descendants never get their own copy.
//
// # A nil Handler is allowed when Children is declared
//
// Handler represents a command's own business logic, so assemble does not
// force a command to supply one just because it has none — a command that
// exists only to hold children (e.g. the root command) has no logic of its
// own. Forcing a placeholder Handler just to satisfy that requirement would
// conflate plumbing with the Handler abstraction's actual meaning. Instead,
// execute leaves RunE nil for such commands, and cobra's own Runnable()
// convention takes over: a command with no Run/RunE is treated as
// non-runnable, and cobra invokes its resolved HelpFunc when the command is
// invoked with no matching subcommand — whatever that HelpFunc does.
//
// # System flag effects fire at parse time, not in RunE
//
// Cobra's non-Runnable fallback (used above) returns flag.ErrHelp before
// cobra ever calls preRun — the step that would invoke PersistentPreRunE.
// Firing a system flag's effect from PersistentPreRunE would therefore
// silently never happen whenever a Handler-less command runs bare, with a
// persistent flag set but no subcommand given. effectValue sidesteps this by
// firing the effect immediately from pflag.Value.Set, which cobra calls while
// parsing flags — well before any Runnable check exists. This makes the
// effect fire regardless of whether the invoked command is ever Runnable,
// and regardless of whether the flag appears before or after a subcommand
// name, since cobra merges persistent flags into every descendant's flag set.
//
// # Implicit vs. explicit system flags
//
// implicitSystemFlags (Sub: SubImplicit) are engine-owned: no command
// declares them, and Build includes them in every tree automatically.
// Explicit system flags (Sub: SubExplicit) are canonical, shared flag
// definitions that individual commands opt into via Definition.Flags, the
// same way a regular command flag is declared — assemble rejects any attempt
// to declare an implicit flag that way, since that's the engine's exclusive
// concern.
package factory
