// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

// Package flags declares the typed flag definitions a command can carry:
// [CommandFlag] and [SystemFlag], built from shared value descriptors —
// [BoolValue], [StringValue], and [StringSliceValue].
//
// # Two flag kinds: who consumes the value
//
// [CommandFlag] and [SystemFlag] both wrap a value descriptor, but differ in
// [FlagKind]: KindCommand flags are read by the command that declared them;
// KindSystem flags instead carry an Effect applied at the infrastructure
// level. [FlagSubcategory] further splits system flags into SubImplicit
// (applied to every command automatically) and SubExplicit (a shared
// definition a command opts into via its own Flags). CommandFlag has no
// subcategory; its behavior is driven by whether Resolver is set instead.
//
// # Value descriptors are generic over ValueType
//
// [BoolValue], [StringValue], and [StringSliceValue] share one mechanism —
// CommandFlag[V] and SystemFlag[V] are generic over [ValueType], a closed
// union of exactly these three pointer types plus the methods they must
// implement. Each provides the same WithShorthand/WithDefault/WithDest
// builder pattern, mutating the receiver in place and returning it for
// chaining at construction time.
//
// # Resolver is an opt-in escape hatch
//
// Most flags write their parsed value straight into Dest(). A flag that
// needs custom handling after parsing — validating, transforming, or
// writing into a different destination type — sets [Resolver] instead.
// [BoolResolver], [StringResolver], and [StringSliceResolver] each live
// beside the value type they apply to, so a resolver's contract stays next
// to the type it resolves.
//
// # Validate guards defaults, not runtime input
//
// [StringValue.Validate] and [StringSliceValue.Validate] reject null bytes
// and control characters, but only in the default value a developer
// hardcodes — never in what a user types at runtime. This catches
// copy-paste mistakes that would corrupt help text or terminal output, not
// malicious CLI input.
package flags
