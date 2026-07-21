// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import (
	"os/exec"
)

// pmDetector is an internal interface for package managers.
type pmDetector interface {
	Detect(lookPath LookPath, o OS) bool
	PackageManager() PackageManager
}

// PkgDetector implements Detector[[]PackageManagerInfo] by orchestrating
// multiple package manager specific detectors.
type PkgDetector struct {
	os        OS
	lookPath  LookPath
	detectors []pmDetector
}

// NewPkgDetector returns a PkgDetector initialized with all supported
// package manager detectors.
func NewPkgDetector(o OS) *PkgDetector {
	return &PkgDetector{
		os:       o,
		lookPath: exec.LookPath,
		detectors: []pmDetector{
			&homebrewDetector{},
			&aptDetector{},
		},
	}
}

// Detect returns a list of all detected system package managers.
func (d *PkgDetector) Detect() []PackageManagerInfo {
	var managers []PackageManagerInfo
	for _, det := range d.detectors {
		if det.Detect(d.lookPath, d.os) {
			managers = append(managers, PackageManagerInfo{
				Name:    det.PackageManager(),
				Present: true,
			})
		}
	}
	return managers
}

// Individual implementations

type homebrewDetector struct{}

func (homebrewDetector) PackageManager() PackageManager { return Homebrew }
func (homebrewDetector) Detect(lp LookPath, _ OS) bool {
	_, err := lp(binBrew)
	return err == nil
}

type aptDetector struct{}

func (aptDetector) PackageManager() PackageManager { return Apt }
func (aptDetector) Detect(lp LookPath, o OS) bool {
	if o != Linux {
		return false
	}
	_, err := lp(binApt)
	return err == nil
}
