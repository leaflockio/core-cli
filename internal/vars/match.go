// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package vars

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

// Comparison controls how strictly a var's resolved value must match when
// Match builds a comparison regex. It only affects vars whose Kind is not
// KindStatic.
type Comparison int

const (
	// Auto derives each var's match mode from its own Volatility: Volatile
	// vars match by pattern, Stable vars match literally. This is the
	// zero value, so an unset Comparison behaves like Auto.
	Auto Comparison = iota
	// Strict forces every non-Static var to match literally, even
	// Volatile ones.
	Strict
	// Loose forces every non-Static var to match by pattern, even Stable
	// ones.
	Loose
)

// usePattern reports whether v should contribute its wildcard pattern
// (true) or its literal value (false) when Match is building a regex
// under comparison.
func (v *Var) usePattern(comparison Comparison) bool {
	if v.kind == KindStatic {
		return false
	}
	switch comparison {
	case Auto:
		return v.volatility == Volatile
	case Strict:
		return false
	case Loose:
		return true
	default:
		return v.volatility == Volatile
	}
}

// Match builds a regex for template under comparison. Each variable
// contributes either its wildcard pattern or its literal value.
func (v *Vars) Match(template string, comparison Comparison) (*regexp.Regexp, error) {
	pattern, err := v.matchSubstitute(template, comparison, map[string]bool{})
	v.clearCalls()
	if err != nil {
		return nil, err
	}
	return regexp.Compile(pattern)
}

// matchSubstitute walks s for {NAME} references the same way substitute
// does, but regex-escapes literal spans and asks matchExpand for each
// name's contribution.
func (v *Vars) matchSubstitute(s string, comparison Comparison, seen map[string]bool) (string, error) {
	matches := namePattern.FindAllStringSubmatchIndex(s, -1)
	if matches == nil {
		return regexp.QuoteMeta(s), nil
	}

	var out strings.Builder
	last := 0
	for _, m := range matches {
		start, end, nameStart, nameEnd := m[0], m[1], m[2], m[3]
		name := strings.ToUpper(s[nameStart:nameEnd])

		fragment, err := v.matchExpand(name, comparison, seen)
		if err != nil {
			return "", err
		}

		out.WriteString(regexp.QuoteMeta(s[last:start]))
		out.WriteString(fragment)
		last = end
	}
	out.WriteString(regexp.QuoteMeta(s[last:]))
	return out.String(), nil
}

// matchExpand contributes name's regex fragment: its wildcard pattern when
// usePattern calls for it, otherwise its literal value — itself walked
// through matchSubstitute, so a nested reference inside that value gets
// its own usePattern treatment rather than being flattened to plain text.
func (v *Vars) matchExpand(name string, comparison Comparison, seen map[string]bool) (string, error) {
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

	if target.usePattern(comparison) {
		return target.pattern, nil
	}

	raw, err := target.resolve(v.app)
	if err != nil {
		return "", err
	}

	seen[name] = true
	defer delete(seen, name)
	return v.matchSubstitute(raw, comparison, seen)
}
