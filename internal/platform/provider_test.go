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

// TestProvider_BaseRef_azureStripsRefsHeadsPrefix is a regression test:
// Azure Pipelines' SYSTEM_PULLREQUEST_TARGETBRANCH is a full ref
// ("refs/heads/main"), unlike every other provider's bare branch name.
// Without stripping the prefix, ResolveBase would build "origin/refs/heads/main",
// not a valid git ref.
func TestProvider_BaseRef_azureStripsRefsHeadsPrefix(t *testing.T) {
	t.Setenv(envAzureTargetBranch, "refs/heads/main")
	ref, _ := ProviderAzure.BaseRef()
	if ref != "main" {
		t.Errorf("ref = %q, want %q", ref, "main")
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
