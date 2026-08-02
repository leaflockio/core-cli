// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags

import (
	"fmt"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

// Rule validates a relationship between a set of flags, using their parsed
// state. Each implementation owns its own meaning and error.
type Rule interface {
	// Check reports an error if the rule is violated. changed reports
	// whether the named flag was set on the command line.
	Check(changed func(name string) bool) error
}

// Exclusive is a Rule requiring at most one of Flags to be set.
type Exclusive struct {
	Flags []Flag
}

func (r Exclusive) Check(changed func(name string) bool) error {
	var set []string
	for _, f := range r.Flags {
		if changed(f.Definition().Meta.Name) {
			set = append(set, flagRef(f))
		}
	}
	if len(set) <= 1 {
		return nil
	}

	all := strings.Join(refsOf(r.Flags), ", ")
	return errs.Caller(errs.FLR001,
		fmt.Sprintf("%s are mutually exclusive", all),
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("more than one was provided: %s", strings.Join(set, ", ")),
			Resolution: fmt.Sprintf("use exactly one of %s", all),
		},
	)
}

// Requires is a Rule enforcing an OR relationship: whenever Flag is set, at
// least one flag in Needs must also be set.
type Requires struct {
	Flag  Flag
	Needs []Flag
}

func (r Requires) Check(changed func(name string) bool) error {
	name := r.Flag.Definition().Meta.Name
	if !changed(name) {
		return nil
	}
	for _, need := range r.Needs {
		if changed(need.Definition().Meta.Name) {
			return nil
		}
	}

	self := flagRef(r.Flag)
	needed := strings.Join(refsOf(r.Needs), ", ")
	requirement, missing, add := needed, needed, needed
	if len(r.Needs) > 1 {
		requirement = "at least one of " + needed
		missing = "any of " + needed
		add = "one of " + needed
	}

	return errs.Caller(errs.FLR002,
		fmt.Sprintf("%s requires %s", self, requirement),
		nil,
		errs.Context{
			Cause:      fmt.Sprintf("%s was provided without %s", self, missing),
			Resolution: fmt.Sprintf("add %s, or remove %s", add, self),
		},
	)
}

// flagRef renders f's canonical reference for a rule-violation message:
// "--name/-s" when a shorthand is registered, "--name" otherwise.
func flagRef(f Flag) string {
	m := f.Definition().Meta
	if s := m.ShortFlag(); s != "" {
		return m.LongFlag() + "/" + s
	}
	return m.LongFlag()
}

// refsOf renders each flag's canonical reference, in order.
func refsOf(fs []Flag) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = flagRef(f)
	}
	return out
}
