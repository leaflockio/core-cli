// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

// namePattern matches a {NAME} reference inside a template.
var namePattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

var (
	errCompositionCycle = errors.New("vars: composition cycle detected")
	errUnknownVariable  = errors.New("vars: unknown variable")
)

// Resolve fills every {NAME} reference in template with its variable's
// value, matching names case-insensitively. A variable's own value may
// itself reference other variables, those are expanded too, recursively,
// until nothing more can be substituted. It errors on an unknown name, a
// PerCall name with no value supplied for this call, or a composition cycle.
//
// Once Resolve returns, every PerCall variable's compute function is
// cleared, whether or not this call used it — a later Resolve call needs
// a fresh SetCall for any PerCall name it references.
func (v *Vars) Resolve(template string) (string, error) {
	result, err := v.substitute(template, map[string]bool{})
	v.clearCalls()
	return result, err
}

// substitute expands every {NAME} in s, tracking names currently being
// expanded via seen to detect composition cycles.
func (v *Vars) substitute(s string, seen map[string]bool) (string, error) {
	matches := namePattern.FindAllStringSubmatchIndex(s, -1)
	if matches == nil {
		return s, nil
	}

	var out strings.Builder
	last := 0
	for _, m := range matches {
		start, end, nameStart, nameEnd := m[0], m[1], m[2], m[3]
		name := strings.ToUpper(s[nameStart:nameEnd])

		value, err := v.expand(name, seen)
		if err != nil {
			return "", err
		}

		out.WriteString(s[last:start])
		out.WriteString(value)
		last = end
	}
	out.WriteString(s[last:])
	return out.String(), nil
}

// expand resolves name's value and recursively expands any {NAME}
// references inside it.
func (v *Vars) expand(name string, seen map[string]bool) (string, error) {
	if seen[name] {
		return "", errs.Caller(errs.VAR002,
			fmt.Sprintf("composition cycle detected at %s", name),
			fmt.Errorf("%w: %s", errCompositionCycle, name),
			errs.Context{
				Cause:      fmt.Sprintf("%s's value expands back to itself, directly or through other variables", name),
				Resolution: "remove the circular reference from the template or var definitions",
			},
		)
	}
	target, ok := v.base[name]
	if !ok {
		return "", errs.Caller(errs.VAR003,
			fmt.Sprintf("unknown variable %s", name),
			fmt.Errorf("%w: %s", errUnknownVariable, name),
			errs.Context{
				Cause: fmt.Sprintf("the template references {%s}, which isn't a known variable", name),
				Resolution: fmt.Sprintf(
					"check {%s} for a typo, or define it as a variable before resolving this template",
					name,
				),
			},
		)
	}

	raw, err := target.resolve(v.app)
	if err != nil {
		return "", err
	}

	seen[name] = true
	defer delete(seen, name)
	return v.substitute(raw, seen)
}

// clearCalls resets every PerCall variable's compute function, so the next
// Resolve call requires a fresh SetCall.
func (v *Vars) clearCalls() {
	for _, target := range v.base {
		if target.kind == KindPerCall {
			target.compute = nil
		}
	}
}
