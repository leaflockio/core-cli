// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package concurrent

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

var (
	errTaskA = errors.New("task a failed")
	errTaskB = errors.New("task b failed")
)

func TestRun_executesAllTasks(t *testing.T) {
	var count int32
	tasks := make([]func() error, 10)
	for i := range tasks {
		tasks[i] = func() error {
			atomic.AddInt32(&count, 1)
			return nil
		}
	}

	if errs := Run(tasks...); errs != nil {
		t.Errorf("expected no errors, got %v", errs)
	}
	if count != 10 {
		t.Errorf("expected all 10 tasks to run, got %d", count)
	}
}

func TestRun_waitsForSlowestTask(t *testing.T) {
	var done int32

	Run(
		func() error {
			time.Sleep(20 * time.Millisecond)
			atomic.StoreInt32(&done, 1)
			return nil
		},
		func() error { return nil },
	)

	if done != 1 {
		t.Error("expected Run to wait for the slower task before returning")
	}
}

func TestRun_collectsAllErrors(t *testing.T) {
	errs := Run(
		func() error { return errTaskA },
		func() error { return nil },
		func() error { return errTaskB },
	)

	if len(errs) != 2 {
		t.Fatalf("expected 2 collected errors, got %v", errs)
	}
	var gotA, gotB bool
	for _, err := range errs {
		switch {
		case errors.Is(err, errTaskA):
			gotA = true
		case errors.Is(err, errTaskB):
			gotB = true
		}
	}
	if !gotA || !gotB {
		t.Errorf("expected both errTaskA and errTaskB in %v", errs)
	}
}

func TestRun_noTasks(t *testing.T) {
	if errs := Run(); errs != nil {
		t.Errorf("expected no errors for zero tasks, got %v", errs)
	}
}
