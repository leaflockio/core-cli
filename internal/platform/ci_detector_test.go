// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"os"
	"testing"
)

func TestCIDetector_Detect(t *testing.T) {
	tests := []struct {
		name     string
		getenv   Getenv
		expected CIInfo
	}{
		{
			name: "GitHub Actions",
			getenv: func(k string) string {
				if k == envGitHubActions {
					return valTrue
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderGitHub, Present: true},
		},
		{
			name: "GitLab CI",
			getenv: func(k string) string {
				if k == envGitLabCI {
					return valTrue
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderGitLab, Present: true},
		},
		{
			name: "Azure Pipelines",
			getenv: func(k string) string {
				if k == envTFBuild {
					return valTrueC
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderAzure, Present: true},
		},
		{
			name: "Bitbucket Pipelines",
			getenv: func(k string) string {
				if k == envBitbucketBuild {
					return "1"
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderBitbucket, Present: true},
		},
		{
			name: "CircleCI",
			getenv: func(k string) string {
				if k == envCircleCI {
					return valTrue
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderCircle, Present: true},
		},
		{
			name: "Jenkins",
			getenv: func(k string) string {
				if k == envJenkinsURL {
					return "http://jenkins.example.com"
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderJenkins, Present: true},
		},
		{
			name: "Generic CI",
			getenv: func(k string) string {
				if k == envCI {
					return valTrue
				}
				return ""
			},
			expected: CIInfo{Provider: ProviderGeneric, Present: true},
		},
		{
			name: "None",
			getenv: func(k string) string {
				return ""
			},
			expected: CIInfo{Provider: ProviderNone, Present: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &CIDetector{
				getenv: tt.getenv,
				detectors: []providerDetector{
					&githubDetector{},
					&gitlabDetector{},
					&azureDetector{},
					&bitbucketDetector{},
					&circleDetector{},
					&jenkinsDetector{},
					&genericCIDetector{},
				},
			}
			got := d.Detect()
			if got.Provider != tt.expected.Provider {
				t.Errorf("CIDetector.Detect().Provider = %v, want %v", got.Provider, tt.expected.Provider)
			}
			if got.Present != tt.expected.Present {
				t.Errorf("CIDetector.Detect().Present = %v, want %v", got.Present, tt.expected.Present)
			}
		})
	}
}

func TestNewCIDetector(t *testing.T) {
	d := NewCIDetector(os.Getenv)
	if len(d.detectors) == 0 {
		t.Error("NewCIDetector initialized with no detectors")
	}
}
