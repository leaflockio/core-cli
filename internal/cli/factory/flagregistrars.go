// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package factory

import "github.com/spf13/pflag"

// boolRegistrar returns a func that registers a bool flag on a pflag.FlagSet,
// using the VarP variant when shorthand or dest is provided.
func boolRegistrar(name, shorthand, usage string, def bool, dest *bool) func(*pflag.FlagSet) {
	return func(fs *pflag.FlagSet) {
		if dest != nil {
			if shorthand != "" {
				fs.BoolVarP(dest, name, shorthand, def, usage)
			} else {
				fs.BoolVar(dest, name, def, usage)
			}
			return
		}
		if shorthand != "" {
			fs.BoolP(name, shorthand, def, usage)
		} else {
			fs.Bool(name, def, usage)
		}
	}
}

// stringRegistrar returns a func that registers a string flag on a pflag.FlagSet,
// using the VarP variant when shorthand or dest is provided.
func stringRegistrar(name, shorthand, usage, def string, dest *string) func(*pflag.FlagSet) {
	return func(fs *pflag.FlagSet) {
		if dest != nil {
			if shorthand != "" {
				fs.StringVarP(dest, name, shorthand, def, usage)
			} else {
				fs.StringVar(dest, name, def, usage)
			}
			return
		}
		if shorthand != "" {
			fs.StringP(name, shorthand, def, usage)
		} else {
			fs.String(name, def, usage)
		}
	}
}

// stringArrayRegistrar returns a func that registers a string-array flag on a
// pflag.FlagSet, using the VarP variant when shorthand or dest is provided.
func stringArrayRegistrar(name, shorthand, usage string, def []string, dest *[]string) func(*pflag.FlagSet) {
	return func(fs *pflag.FlagSet) {
		if dest != nil {
			if shorthand != "" {
				fs.StringArrayVarP(dest, name, shorthand, def, usage)
			} else {
				fs.StringArrayVar(dest, name, def, usage)
			}
			return
		}
		if shorthand != "" {
			fs.StringArrayP(name, shorthand, def, usage)
		} else {
			fs.StringArray(name, def, usage)
		}
	}
}
