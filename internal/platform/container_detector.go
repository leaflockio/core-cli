// Copyright 2026 LeafLock. All rights reserved.
//
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package platform

import "os"

// containerSubDetector is an internal interface for specific container types.
type containerSubDetector interface {
	Detect(stat Stat) bool
	Type() ContainerType
}

// ContainerDetector implements Detector[ContainerInfo] by orchestrating
// multiple container-specific detectors.
type ContainerDetector struct {
	stat      Stat
	detectors []containerSubDetector
}

// NewContainerDetector returns a ContainerDetector initialized with all
// supported container detectors.
func NewContainerDetector() *ContainerDetector {
	return &ContainerDetector{
		stat: os.Stat,
		detectors: []containerSubDetector{
			&dockerDetector{},
			&k8sDetector{},
		},
	}
}

// Detect iterates through registered detectors and returns ContainerInfo.
func (d *ContainerDetector) Detect() ContainerInfo {
	for _, det := range d.detectors {
		if det.Detect(d.stat) {
			return ContainerInfo{
				Type:    det.Type(),
				Present: true,
			}
		}
	}
	return ContainerInfo{
		Type:    ContainerNone,
		Present: false,
	}
}

// Individual implementations

type dockerDetector struct{}

func (dockerDetector) Type() ContainerType { return ContainerDocker }
func (dockerDetector) Detect(s Stat) bool {
	_, err := s(PathDockerEnv)
	return err == nil
}

type k8sDetector struct{}

func (k8sDetector) Type() ContainerType { return ContainerK8s }
func (k8sDetector) Detect(s Stat) bool {
	_, err := s(PathK8sSecret)
	return err == nil
}
