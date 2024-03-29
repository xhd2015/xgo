//go:build go1.18
// +build go1.18

package mock_generic

import (
	"fmt"
	"testing"
)

type Formatter[T any] struct {
	prefix T
}

func (c *Formatter[T]) Format(v T) string {
	return fmt.Sprintf("%v: %v", c.prefix, v)
}

func TestMockGenericFunc(t *testing.T) {
	formatter := &Formatter[int]{prefix: 1}

	a := formatter.Format
	a(0)
}
