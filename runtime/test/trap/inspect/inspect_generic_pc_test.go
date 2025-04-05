//go:build go1.18
// +build go1.18

package inspect

import (
	"fmt"
)

func ToString[T any](v T) string {
	return fmt.Sprint(v)
}
