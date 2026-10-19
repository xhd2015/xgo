//go:build go1.18
// +build go1.18

package trap_guard

func First[T any](items ...T) (first T, ok bool) {
	if len(items) == 0 {
		return first, false
	}
	return items[0], true
}
