// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import "os"

// Provider identifies the CI/CD platform the process is running on.
type Provider int

const (
	ProviderNone      Provider = iota // not running in CI
	ProviderGitHub                    // GitHub Actions
	ProviderGitLab                    // GitLab CI
	ProviderAzure                     // Azure DevOps Pipelines
	ProviderBitbucket                 // Bitbucket Pipelines
	ProviderCircle                    // CircleCI
	ProviderJenkins                   // Jenkins
	ProviderGeneric                   // CI=true but provider unknown
)

// CI provider display names.
const (
	nameGitHub    = "github-actions"
	nameGitLab    = "gitlab-ci"
	nameAzure     = "azure-devops"
	nameBitbucket = "bitbucket-pipelines"
	nameCircle    = "circleci"
	nameJenkins   = "jenkins"
	nameGeneric   = "ci"
	nameNone      = "none"
)

// String returns the display name of the CI provider.
func (p Provider) String() string {
	switch p {
	case ProviderGitHub:
		return nameGitHub
	case ProviderGitLab:
		return nameGitLab
	case ProviderAzure:
		return nameAzure
	case ProviderBitbucket:
		return nameBitbucket
	case ProviderCircle:
		return nameCircle
	case ProviderJenkins:
		return nameJenkins
	case ProviderGeneric:
		return nameGeneric
	case ProviderNone:
		return nameNone
	default:
		return nameNone
	}
}

// IsCI reports whether any known CI environment was detected.
func (p Provider) IsCI() bool { return p != ProviderNone }

// BaseRef returns the PR target branch ref and merge-base SHA for the provider,
// read from the CI environment variables set by each platform.
// Returns empty strings when not in a PR context or provider is unknown.
func (p Provider) BaseRef() (string, string) {
	switch p {
	case ProviderGitHub:
		return os.Getenv(envGitHubBaseRef), os.Getenv(envGitHubBaseSHA)
	case ProviderGitLab:
		return os.Getenv(envGitLabTargetBranch), os.Getenv(envGitLabBaseSHA)
	case ProviderAzure:
		return os.Getenv(envAzureTargetBranch), ""
	case ProviderBitbucket:
		return os.Getenv(envBitbucketTargetBranch), ""
	case ProviderNone, ProviderCircle, ProviderJenkins, ProviderGeneric:
		return "", ""
	default:
		return "", ""
	}
}

// ResolveBase returns the git ref to use as the PR base. Precedence:
// --base flag > CI merge-base SHA > CI target branch ref > origin/main fallback.
func (p Provider) ResolveBase(flagBase string) string {
	if flagBase != "" {
		return flagBase
	}
	ref, sha := p.BaseRef()
	if sha != "" {
		return sha
	}
	if ref != "" {
		return DefaultRemote + "/" + ref
	}
	return DefaultRemote + "/" + DefaultBranch
}
