//go:build go1.18
// +build go1.18

package mock_generic

import (
	"fmt"
)

func ToString[T any](v T) string {
	return fmt.Sprint(v)
}

type Formatter[T any] struct {
	prefix T
}

func (c *Formatter[T]) Format(v T) string {
	return fmt.Sprintf("%v: %v", c.prefix, v)
}
