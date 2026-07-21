// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package config

import (
	"errors"
	"testing"
)

func TestEnvFromString(t *testing.T) {
	tests := []struct {
		input   string
		want    Env
		wantErr bool
	}{
		{"dev", EnvDev, false},
		{"test", EnvTest, false},
		{"prod", EnvProd, false},
		{"DEV", EnvDev, false},
		{"TEST", EnvTest, false},
		{"PROD", EnvProd, false},
		{"", EnvDev, false},
		{"staging", "", true},
		{"production", "", true},
		{"Dev", EnvDev, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := EnvFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EnvFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("EnvFromString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEnvFromEnvVar(t *testing.T) {
	t.Setenv(envVarEnv, "test")

	got, err := EnvFromEnvVar()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != EnvTest {
		t.Errorf("got %q, want %q", got, EnvTest)
	}
}

func TestEnvFromEnvVar_empty(t *testing.T) {
	t.Setenv(envVarEnv, "")

	got, err := EnvFromEnvVar()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != EnvDev {
		t.Errorf("got %q, want %q (empty should default to dev)", got, EnvDev)
	}
}

func TestEnvFromEnvVar_invalid(t *testing.T) {
	t.Setenv(envVarEnv, "staging")

	_, err := EnvFromEnvVar()
	if err == nil {
		t.Fatal("expected error for unknown env, got nil")
	}
	if !errors.Is(err, ErrUnknownEnv) {
		t.Errorf("expected errors.Is(err, ErrUnknownEnv), got %v", err)
	}
}

func TestEnvResolve_envVarWins(t *testing.T) {
	t.Setenv(envVarEnv, "prod")

	got, err := EnvResolve("dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != EnvProd {
		t.Errorf("got %q, want %q (LEAF_ENV should win over buildDefault)", got, EnvProd)
	}
}

func TestEnvResolve_fallsBackToBuildDefault(t *testing.T) {
	t.Setenv(envVarEnv, "")

	got, err := EnvResolve("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != EnvTest {
		t.Errorf("got %q, want %q (should use buildDefault)", got, EnvTest)
	}
}

func TestEnvResolve_invalidEnvVar(t *testing.T) {
	t.Setenv(envVarEnv, "staging")

	_, err := EnvResolve("dev")
	if err == nil {
		t.Fatal("expected error for invalid LEAF_ENV, got nil")
	}
	if !errors.Is(err, ErrUnknownEnv) {
		t.Errorf("expected errors.Is(err, ErrUnknownEnv), got %v", err)
	}
}

func TestEnvResolve_invalidBuildDefault(t *testing.T) {
	t.Setenv(envVarEnv, "")

	_, err := EnvResolve("staging")
	if err == nil {
		t.Fatal("expected error for invalid buildDefault, got nil")
	}
	if !errors.Is(err, ErrUnknownEnv) {
		t.Errorf("expected errors.Is(err, ErrUnknownEnv), got %v", err)
	}
}

func TestEnvString(t *testing.T) {
	tests := []struct {
		env  Env
		want string
	}{
		{EnvDev, "dev"},
		{EnvTest, "test"},
		{EnvProd, "prod"},
	}

	for _, tt := range tests {
		if got := tt.env.String(); got != tt.want {
			t.Errorf("%v.String() = %q, want %q", tt.env, got, tt.want)
		}
	}
}
