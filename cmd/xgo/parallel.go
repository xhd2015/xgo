package main

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// forEachParallel calls f(i) for every i in [0, n) on up to GOMAXPROCS
// goroutines, handing out indices in order, and returns when all are done.
func forEachParallel(n int, f func(i int)) {
	workers := runtime.GOMAXPROCS(0)
	if workers > n {
		workers = n
	}
	if workers <= 1 {
		for i := 0; i < n; i++ {
			f(i)
		}
		return
	}
	var next int64 = -1
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				i := int(atomic.AddInt64(&next, 1))
				if i >= n {
					return
				}
				f(i)
			}
		}()
	}
	wg.Wait()
}
