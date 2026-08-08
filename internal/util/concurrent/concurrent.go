// Copyright 2026 LeafLock. All rights reserved.
// This source code is proprietary and confidential.
// Unauthorized copying, modification, distribution, or use of this
// software, via any medium, is strictly prohibited without prior
// written permission from LeafLock.

package concurrent

import "sync"

// Run runs each task in its own goroutine and waits for all of them to
// finish, collecting the errors they return. The returned slice has one
// entry per task that returned a non-nil error, in no guaranteed order —
// nil when every task succeeded.
func Run(tasks ...func() error) []error {
	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func() {
			defer wg.Done()
			if err := task(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return errs
}
