// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

// providerDetector is an internal interface for specific CI providers.
type providerDetector interface {
	Detect(getenv Getenv) bool
	Provider() Provider
}

// CIDetector implements Detector[CIInfo] by orchestrating multiple
// provider-specific detectors.
type CIDetector struct {
	getenv    Getenv
	detectors []providerDetector
}

// NewCIDetector returns a CIDetector initialized with all supported
// CI provider detectors.
func NewCIDetector(getenv Getenv) *CIDetector {
	return &CIDetector{
		getenv: getenv,
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
}

// Detect iterates through registered detectors and returns CIInfo.
func (d *CIDetector) Detect() CIInfo {
	for _, det := range d.detectors {
		if det.Detect(d.getenv) {
			p := det.Provider()
			return CIInfo{
				Provider: p,
				Present:  p.IsCI(),
			}
		}
	}
	return CIInfo{
		Provider: ProviderNone,
		Present:  false,
	}
}

// Individual implementations

type githubDetector struct{}

func (githubDetector) Provider() Provider   { return ProviderGitHub }
func (githubDetector) Detect(g Getenv) bool { return g(envGitHubActions) == valTrue }

type gitlabDetector struct{}

func (gitlabDetector) Provider() Provider   { return ProviderGitLab }
func (gitlabDetector) Detect(g Getenv) bool { return g(envGitLabCI) == valTrue }

type azureDetector struct{}

func (azureDetector) Provider() Provider   { return ProviderAzure }
func (azureDetector) Detect(g Getenv) bool { return g(envTFBuild) == valTrueC }

type bitbucketDetector struct{}

func (bitbucketDetector) Provider() Provider   { return ProviderBitbucket }
func (bitbucketDetector) Detect(g Getenv) bool { return g(envBitbucketBuild) != "" }

type circleDetector struct{}

func (circleDetector) Provider() Provider   { return ProviderCircle }
func (circleDetector) Detect(g Getenv) bool { return g(envCircleCI) == valTrue }

type jenkinsDetector struct{}

func (jenkinsDetector) Provider() Provider   { return ProviderJenkins }
func (jenkinsDetector) Detect(g Getenv) bool { return g(envJenkinsURL) != "" }

type genericCIDetector struct{}

func (genericCIDetector) Provider() Provider   { return ProviderGeneric }
func (genericCIDetector) Detect(g Getenv) bool { return g(envCI) == valTrue }
