// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import "github.com/spf13/pflag"

// effectValue wraps a pflag.Value so effect fires immediately after a
// successful Set, during flag parsing — before cobra's RunE/Runnable exist.
type effectValue struct {
	pflag.Value
	effect func()
}

func (v *effectValue) Set(s string) error {
	if err := v.Value.Set(s); err != nil {
		return err
	}
	v.effect()
	return nil
}

// IsBoolFlag delegates to the wrapped Value so bool flags still parse
// without an explicit value.
func (v *effectValue) IsBoolFlag() bool {
	bf, ok := v.Value.(interface{ IsBoolFlag() bool })
	return ok && bf.IsBoolFlag()
}

// wrapWithEffect looks up name on fs and wraps its Value so effect fires on
// every successful Set. Must be called after the flag has been registered.
func wrapWithEffect(fs *pflag.FlagSet, name string, effect func()) {
	f := fs.Lookup(name)
	f.Value = &effectValue{Value: f.Value, effect: effect}
}
