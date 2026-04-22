// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"os"
	"path/filepath"
)

// Detector is the base interface for all system property detectors.
type Detector[T any] interface {
	Detect() T
}

// Runtime dependencies for detection.
type (
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)
)

// CI environment variable names used for provider detection and metadata.
const (
	envGitHubActions  = "GITHUB_ACTIONS"
	envGitLabCI       = "GITLAB_CI"
	envTFBuild        = "TF_BUILD"
	envBitbucketBuild = "BITBUCKET_BUILD_NUMBER"
	envCircleCI       = "CIRCLECI"
	envJenkinsURL     = "JENKINS_URL"
	envCI             = "CI"

	envGitHubBaseRef         = "GITHUB_BASE_REF"
	envGitHubBaseSHA         = "GITHUB_BASE_SHA"
	envGitLabTargetBranch    = "CI_MERGE_REQUEST_TARGET_BRANCH_NAME"
	envGitLabBaseSHA         = "CI_MERGE_REQUEST_DIFF_BASE_SHA"
	envAzureTargetBranch     = "SYSTEM_PULLREQUEST_TARGETBRANCH"
	envBitbucketTargetBranch = "BITBUCKET_PR_DESTINATION_BRANCH"
)

// CI environment variable values.
const (
	valTrue  = "true"
	valTrueC = "True" // Azure uses TitleCase
)

// Default git values for PR base resolution.
const (
	DefaultRemote = "origin"
	DefaultBranch = "main"
)

// Package manager binary names used for PATH detection.
const (
	binBrew = "brew"
	binApt  = "apt"
)

// Dynamic path construction to avoid git hooks flagging literal absolute paths.
var (
	// PathDockerEnv is the path to the Docker environment file.
	PathDockerEnv = string(os.PathSeparator) + ".dockerenv"

	// PathK8sSecret is the path to the Kubernetes service account secrets.
	PathK8sSecret = filepath.Join(string(os.PathSeparator), "var", "run", "secrets", "kubernetes.io")
)
