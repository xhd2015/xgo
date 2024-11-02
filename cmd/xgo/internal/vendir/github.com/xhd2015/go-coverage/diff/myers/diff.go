// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package myers implements the Myers diff algorithm.
package myers

import (
	"runtime"
	"strings"
	"sync"
	"time"
)

// Sources:
// https://blog.jcoglan.com/2017/02/17/the-myers-diff-algorithm-part-3/
// https://www.codeproject.com/Articles/42279/%2FArticles%2F42279%2FInvestigating-Myers-diff-algorithm-Part-1-of-2

type Operation operation

type operation struct {
	Kind    OpKind
	Content []string // content from b
	I1, I2  int      // indices of the line in a
	J1      int      // indices of the line in b, J2 implied by len(Content)
}

// operations returns the list of operations to convert a into b, consolidating
// operations for multiple lines and not including equal lines.
func operations(a, b []string) []*operation {
	return operationsComplex(a, b, nil, nil)
}

var alloc arrAlloctor

func newArrNaive(size int) []int {
	return make([]int, size)
}
func newArrPooled(size int) []int {
	return alloc.newArr(size)
}
func returnArrPooled(size int, arr []int) {
	alloc.returnArr(size, arr)
}

type arrAlloctor struct {
	mutex sync.Mutex
	pool  [][][]int
}

func (c *arrAlloctor) newArr(size int) []int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	targetLen := size + 1
	n := len(c.pool)
	if targetLen > n {
		c.pool = append(c.pool, make([][][]int, targetLen-n)...)
	}
	p := c.pool[size]
	var arr []int
	if len(p) == 0 {
		arr = make([]int, size)
		p = append(p, arr)
	} else {
		arr = p[len(p)-1]
		p = p[:len(p)-1]
		// reset
		for i := 0; i < len(arr); i++ {
			arr[i] = 0
		}
	}
	c.pool[size] = p
	return arr
}

func (c *arrAlloctor) returnArr(size int, arr []int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if size >= len(c.pool) {
		return
	}
	c.pool[size] = append(c.pool[size], arr)
}

// for `onUpdate`, if newLineEnd - newLineStart = 0, then it is a deletion. Otherwise an update.
// NOTE: newLineEnd,oldLineEnd is exclusive.
func operationsComplex(a, b []string, onSame func(oldLine, newLine int), onUpdate func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int)) []*operation {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	// newArrPooled has bug

	usePooled := false // usePooled no improve
	useGc := false     // useGc no improve
	useSleep := false  // useSleep no improve

	if useSleep {
		time.Sleep(10 * time.Millisecond)
	}

	// // space complexity:  O((N+M)^2)
	var trace [][]int
	var offset int

	if useGc {
		runtime.GC()
	}

	if usePooled {
		trace, offset = shortestEditSequence(a, b, newArrPooled)
		defer func() {
			for _, arr := range trace {
				returnArrPooled(len(arr), arr)
			}
		}()
	} else {
		// newArrNaive consumes too much memroy
		trace, offset = shortestEditSequence(a, b, newArrNaive)
	}

	snakes := backtrack(trace, len(a), len(b), offset)

	M, N := len(a), len(b)

	var i int
	solution := make([]*operation, len(a)+len(b))

	add := func(op *operation, i2, j2 int) {
		if op == nil {
			return
		}
		op.I2 = i2
		if op.Kind == Insert {
			op.Content = b[op.J1:j2]
		}
		solution[i] = op
		i++
	}
	// x: line track of old content
	// y: line track of new content
	x, y := 0, 0
	for _, snake := range snakes {
		if len(snake) < 2 {
			continue
		}
		oldLineStart := x
		var op *operation
		// delete (horizontal)
		for snake[0]-snake[1] > x-y {
			if op == nil {
				op = &operation{
					Kind: Delete,
					I1:   x,
					J1:   y,
				}
			}
			x++
			if x == M {
				break
			}
		}
		oldLineEnd := x
		add(op, x, y)
		op = nil

		newLineStart := y
		// insert (vertical)
		for snake[0]-snake[1] < x-y {
			if op == nil {
				op = &operation{
					Kind: Insert,
					I1:   x,
					J1:   y,
				}
			}
			y++
		}

		newLineEnd := y
		add(op, x, y)

		if onUpdate != nil {
			onUpdate(oldLineStart, oldLineEnd, newLineStart, newLineEnd)
		}

		op = nil
		// equal (diagonal)
		for x < snake[0] {
			// NOTE: there are chances where x>=M && y>=N
			if onSame != nil && x < M && y < N {
				onSame(x, y)
			}

			x++
			y++
		}
		if x >= M && y >= N {
			break
		}
	}
	return solution[:i]
}

// backtrack uses the trace for the edit sequence computation and returns the
// "snakes" that make up the solution. A "snake" is a single deletion or
// insertion followed by zero or diagonals.
func backtrack(trace [][]int, x, y, offset int) [][]int {
	snakes := make([][]int, len(trace))
	d := len(trace) - 1
	for ; x > 0 && y > 0 && d > 0; d-- {
		V := trace[d]
		if len(V) == 0 {
			continue
		}
		snakes[d] = []int{x, y}

		k := x - y

		var kPrev int
		if k == -d || (k != d && V[k-1+offset] < V[k+1+offset]) {
			kPrev = k + 1
		} else {
			kPrev = k - 1
		}

		x = V[kPrev+offset]
		y = x - kPrev
	}
	if x < 0 || y < 0 {
		return snakes
	}
	snakes[d] = []int{x, y}
	return snakes
}

// shortestEditSequence returns the shortest edit sequence that converts a into b.
func shortestEditSequence(a, b []string, newArr func(size int) []int) ([][]int, int) {
	M, N := len(a), len(b)
	V := newArr(2*(N+M) + 1)
	offset := N + M
	trace := make([][]int, N+M+1)

	// Iterate through the maximum possible length of the SES (N+M).
	for d := 0; d <= N+M; d++ {
		copyV := newArr(len(V))
		// k lines are represented by the equation y = x - k. We move in
		// increments of 2 because end points for even d are on even k lines.
		for k := -d; k <= d; k += 2 {
			// At each point, we either go down or to the right. We go down if
			// k == -d, and we go to the right if k == d. We also prioritize
			// the maximum x value, because we prefer deletions to insertions.
			var x int
			if k == -d || (k != d && V[k-1+offset] < V[k+1+offset]) {
				x = V[k+1+offset] // down
			} else {
				x = V[k-1+offset] + 1 // right
			}

			y := x - k

			// Diagonal moves while we have equal contents.
			for x < M && y < N && a[x] == b[y] {
				x++
				y++
			}

			V[k+offset] = x

			// Return if we've exceeded the maximum values.
			if x == M && y == N {
				// Makes sure to save the state of the array before returning.
				copy(copyV, V)
				trace[d] = copyV
				return trace, offset
			}
		}

		// Save the state of the array.
		copy(copyV, V)
		trace[d] = copyV
	}
	return nil, 0
}

func splitLines(text string) []string {
	lines := strings.SplitAfter(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
