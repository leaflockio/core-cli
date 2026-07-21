// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"testing"
)

// --- Provider.String ---

func TestProvider_String(t *testing.T) {
	cases := []struct {
		p    Provider
		want string
	}{
		{ProviderNone, "none"},
		{ProviderGitHub, "github-actions"},
		{ProviderGitLab, "gitlab-ci"},
		{ProviderAzure, "azure-devops"},
		{ProviderBitbucket, "bitbucket-pipelines"},
		{ProviderCircle, "circleci"},
		{ProviderJenkins, "jenkins"},
		{ProviderGeneric, "ci"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("Provider(%d).String() = %q, want %q", tc.p, got, tc.want)
		}
	}
}

// --- Provider.BaseRef ---

func TestProvider_BaseRef_gitHub(t *testing.T) {
	t.Setenv(envGitHubBaseRef, "main")
	t.Setenv(envGitHubBaseSHA, "abc123")
	ref, sha := ProviderGitHub.BaseRef()
	if ref != "main" {
		t.Errorf("ref = %q, want %q", ref, "main")
	}
	if sha != "abc123" {
		t.Errorf("sha = %q, want %q", sha, "abc123")
	}
}

func TestProvider_BaseRef_gitLab(t *testing.T) {
	t.Setenv(envGitLabTargetBranch, "develop")
	t.Setenv(envGitLabBaseSHA, "def456")
	ref, sha := ProviderGitLab.BaseRef()
	if ref != "develop" {
		t.Errorf("ref = %q, want %q", ref, "develop")
	}
	if sha != "def456" {
		t.Errorf("sha = %q, want %q", sha, "def456")
	}
}

func TestProvider_BaseRef_azure(t *testing.T) {
	t.Setenv(envAzureTargetBranch, "main")
	ref, sha := ProviderAzure.BaseRef()
	if ref != "main" {
		t.Errorf("ref = %q, want %q", ref, "main")
	}
	if sha != "" {
		t.Errorf("sha = %q, want empty", sha)
	}
}

func TestProvider_BaseRef_bitbucket(t *testing.T) {
	t.Setenv(envBitbucketTargetBranch, "main")
	ref, sha := ProviderBitbucket.BaseRef()
	if ref != "main" {
		t.Errorf("ref = %q, want %q", ref, "main")
	}
	if sha != "" {
		t.Errorf("sha = %q, want empty", sha)
	}
}

func TestProvider_BaseRef_default(t *testing.T) {
	ref, sha := ProviderNone.BaseRef()
	if ref != "" || sha != "" {
		t.Errorf("BaseRef() = (%q, %q), want empty strings", ref, sha)
	}
}

// --- Provider.ResolveBase ---

func TestProvider_ResolveBase_flagTakesPrecedence(t *testing.T) {
	t.Setenv(envGitHubBaseRef, "main")
	t.Setenv(envGitHubBaseSHA, "abc123")
	got := ProviderGitHub.ResolveBase("my-branch")
	if got != "my-branch" {
		t.Errorf("ResolveBase with flag = %q, want %q", got, "my-branch")
	}
}

func TestProvider_ResolveBase_shaTakesPrecedenceOverRef(t *testing.T) {
	t.Setenv(envGitHubBaseRef, "main")
	t.Setenv(envGitHubBaseSHA, "abc123")
	got := ProviderGitHub.ResolveBase("")
	if got != "abc123" {
		t.Errorf("ResolveBase with sha = %q, want %q", got, "abc123")
	}
}

func TestProvider_ResolveBase_refFallback(t *testing.T) {
	t.Setenv(envGitHubBaseRef, "main")
	t.Setenv(envGitHubBaseSHA, "")
	got := ProviderGitHub.ResolveBase("")
	if got != DefaultRemote+"/main" {
		t.Errorf("ResolveBase with ref = %q, want %q", got, DefaultRemote+"/main")
	}
}

func TestProvider_ResolveBase_originMainFallback(t *testing.T) {
	got := ProviderNone.ResolveBase("")
	if got != DefaultRemote+"/"+DefaultBranch {
		t.Errorf("ResolveBase fallback = %q, want %q", got, DefaultRemote+"/"+DefaultBranch)
	}
}
