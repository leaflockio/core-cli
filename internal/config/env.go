// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/leaflockio/core-cli/internal/errs"
)

const errEnvMsgFmt = "%q is not a valid environment: must be one of %s, %s, %s"

// Env represents a runtime environment profile.
type Env string

const (
	EnvDev  Env = "dev"
	EnvTest Env = "test"
	EnvProd Env = "prod"
)

func (e Env) String() string {
	return string(e)
}

// EnvFromString parses s into an Env. Empty string defaults to EnvDev.
func EnvFromString(s string) (Env, error) {
	switch e := Env(strings.ToLower(s)); e {
	case EnvDev, EnvTest, EnvProd:
		return e, nil
	case "":
		return EnvDev, nil
	default:
		return "", errs.Caller(
			errs.ENV001,
			fmt.Sprintf(errEnvMsgFmt, s, EnvDev, EnvTest, EnvProd),
			ErrUnknownEnv,
			errs.Context{
				Cause:      fmt.Sprintf("%q does not match any known environment", s),
				Resolution: fmt.Sprintf("must be one of: %s, %s, %s", EnvDev, EnvTest, EnvProd),
			},
		)
	}
}

// EnvResolve returns the active Env by checking LEAF_ENV first and falling
// back to buildDefault when it is unset.
func EnvResolve(buildDefault string) (Env, error) {
	if os.Getenv(envVarEnv) != "" {
		return EnvFromEnvVar()
	}
	return EnvFromString(buildDefault)
}

// EnvFromEnvVar reads the LEAF_ENV environment variable and returns the corresponding Env.
func EnvFromEnvVar() (Env, error) {
	s := os.Getenv(envVarEnv)
	e, err := EnvFromString(s)
	if err != nil {
		return "", errs.Caller(
			errs.ENV001,
			fmt.Sprintf(errEnvMsgFmt, s, EnvDev, EnvTest, EnvProd),
			ErrUnknownEnv,
			errs.Context{
				Cause:      fmt.Sprintf("%s is set to %q which does not match any known environment", envVarEnv, s),
				Resolution: fmt.Sprintf("set %s to one of: %s, %s, %s", envVarEnv, EnvDev, EnvTest, EnvProd),
			},
		)
	}
	return e, nil
}
