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

type mockFileInfo struct {
	os.FileInfo
}

func TestContainerDetector_Detect(t *testing.T) {
	tests := []struct {
		name     string
		stat     Stat
		expected ContainerInfo
	}{
		{
			name: "Docker",
			stat: func(p string) (os.FileInfo, error) {
				if p == PathDockerEnv {
					return &mockFileInfo{}, nil
				}
				return nil, os.ErrNotExist
			},
			expected: ContainerInfo{Type: ContainerDocker, Present: true},
		},
		{
			name: "Kubernetes",
			stat: func(p string) (os.FileInfo, error) {
				if p == PathK8sSecret {
					return &mockFileInfo{}, nil
				}
				return nil, os.ErrNotExist
			},
			expected: ContainerInfo{Type: ContainerK8s, Present: true},
		},
		{
			name: "None",
			stat: func(p string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			expected: ContainerInfo{Type: ContainerNone, Present: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ContainerDetector{
				stat: tt.stat,
				detectors: []containerSubDetector{
					&dockerDetector{},
					&k8sDetector{},
				},
			}
			got := d.Detect()
			if got.Type != tt.expected.Type {
				t.Errorf("ContainerDetector.Detect().Type = %v, want %v", got.Type, tt.expected.Type)
			}
			if got.Present != tt.expected.Present {
				t.Errorf("ContainerDetector.Detect().Present = %v, want %v", got.Present, tt.expected.Present)
			}
		})
	}
}
