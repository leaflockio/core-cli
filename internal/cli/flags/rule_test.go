// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package flags_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/leaflockio/core-cli/internal/cli/flags"
	"github.com/leaflockio/core-cli/internal/errs"
)

// changedFrom returns a Rule.Check "changed" closure reporting true for
// exactly the given flag names.
func changedFrom(set ...string) func(name string) bool {
	m := make(map[string]bool, len(set))
	for _, s := range set {
		m[s] = true
	}
	return func(name string) bool { return m[name] }
}

// boolFlag returns a bare bool CommandFlag with the given name.
func boolFlag(name string) flags.Flag {
	return flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool(name, "")}
}

// asRuleError extracts the *errs.Error a Rule.Check returned.
func asRuleError(t *testing.T, err error) *errs.Error {
	t.Helper()
	var e *errs.Error
	if !errors.As(err, &e) {
		t.Fatalf("error = %v, want an *errs.Error", err)
	}
	return e
}

func TestExclusive_Check_nil_when_none_set(t *testing.T) {
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), boolFlag("staged")}}
	if err := r.Check(changedFrom()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExclusive_Check_nil_when_one_set(t *testing.T) {
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), boolFlag("staged")}}
	if err := r.Check(changedFrom("all")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExclusive_Check_error_when_two_set(t *testing.T) {
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), boolFlag("staged"), boolFlag("pr")}}
	e := asRuleError(t, r.Check(changedFrom("all", "staged")))

	if e.Code != errs.FLR001 {
		t.Errorf("Code = %q, want %q", e.Code, errs.FLR001)
	}
	if len(e.Contexts) == 0 {
		t.Fatal("want at least one Context")
	}
	cause := e.Contexts[0].Cause
	if !strings.Contains(cause, "--all") || !strings.Contains(cause, "--staged") {
		t.Errorf("cause = %q, want it to name both conflicting flags", cause)
	}
	if strings.Contains(cause, "--pr") {
		t.Errorf("cause = %q, want it to not name --pr, which was not set", cause)
	}
}

func TestExclusive_Check_message_names_every_group_member(t *testing.T) {
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), boolFlag("staged"), boolFlag("pr")}}
	e := asRuleError(t, r.Check(changedFrom("all", "staged")))

	for _, want := range []string{"--all", "--staged", "--pr"} {
		if !strings.Contains(e.Message, want) {
			t.Errorf("Message = %q, want it to name every flag in the group, missing %q", e.Message, want)
		}
	}
}

func TestExclusive_Check_error_when_all_set(t *testing.T) {
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), boolFlag("staged"), boolFlag("pr")}}
	if err := r.Check(changedFrom("all", "staged", "pr")); err == nil {
		t.Error("expected error when every flag in the group is set")
	}
}

func TestExclusive_Check_references_shorthand_when_registered(t *testing.T) {
	pr := flags.CommandFlag[*flags.BoolValue]{Value: flags.Bool("pr", "").WithShorthand("p")}
	r := flags.Exclusive{Flags: []flags.Flag{boolFlag("all"), pr}}
	e := asRuleError(t, r.Check(changedFrom("all", "pr")))

	if !strings.Contains(e.Message, "--pr/-p") {
		t.Errorf("Message = %q, want it to reference --pr/-p", e.Message)
	}
}

func TestRequires_Check_nil_when_flag_not_set(t *testing.T) {
	r := flags.Requires{Flag: boolFlag("base"), Needs: []flags.Flag{boolFlag("pr")}}
	if err := r.Check(changedFrom()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequires_Check_nil_when_need_also_set(t *testing.T) {
	r := flags.Requires{Flag: boolFlag("base"), Needs: []flags.Flag{boolFlag("pr")}}
	if err := r.Check(changedFrom("base", "pr")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequires_Check_error_when_set_without_need(t *testing.T) {
	r := flags.Requires{Flag: boolFlag("base"), Needs: []flags.Flag{boolFlag("pr")}}
	e := asRuleError(t, r.Check(changedFrom("base")))

	if e.Code != errs.FLR002 {
		t.Errorf("Code = %q, want %q", e.Code, errs.FLR002)
	}
	if e.Message != "--base requires --pr" {
		t.Errorf("Message = %q, want %q", e.Message, "--base requires --pr")
	}
	if len(e.Contexts) == 0 {
		t.Fatal("want at least one Context")
	}
	if got := e.Contexts[0].Cause; got != "--base was provided without --pr" {
		t.Errorf("Cause = %q, want %q", got, "--base was provided without --pr")
	}
	if got := e.Contexts[0].Resolution; got != "add --pr, or remove --base" {
		t.Errorf("Resolution = %q, want %q", got, "add --pr, or remove --base")
	}
}

func TestRequires_Check_nil_when_any_one_of_multiple_needs_set(t *testing.T) {
	r := flags.Requires{Flag: boolFlag("base"), Needs: []flags.Flag{boolFlag("pr"), boolFlag("all")}}
	if err := r.Check(changedFrom("base", "all")); err != nil {
		t.Errorf("unexpected error when one of several Needs is set: %v", err)
	}
}

func TestRequires_Check_message_for_multiple_needs_reads_as_or(t *testing.T) {
	r := flags.Requires{Flag: boolFlag("base"), Needs: []flags.Flag{boolFlag("pr"), boolFlag("all")}}
	e := asRuleError(t, r.Check(changedFrom("base")))

	if e.Message != "--base requires at least one of --pr, --all" {
		t.Errorf("Message = %q, want %q", e.Message, "--base requires at least one of --pr, --all")
	}
	if got := e.Contexts[0].Cause; got != "--base was provided without any of --pr, --all" {
		t.Errorf("Cause = %q, want %q", got, "--base was provided without any of --pr, --all")
	}
	if got := e.Contexts[0].Resolution; got != "add one of --pr, --all, or remove --base" {
		t.Errorf("Resolution = %q, want %q", got, "add one of --pr, --all, or remove --base")
	}
}
