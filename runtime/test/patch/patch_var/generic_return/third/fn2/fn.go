//go:build go1.18
// +build go1.18

package fn2

type Tree2[Q any, V any] struct {
	Root Q
	Data V
}

func MustBuild2[Q any, V any](root Q, data V) *Tree2[Q, V] {
	return &Tree2[Q, V]{
		Root: root,
		Data: data,
	}
}
