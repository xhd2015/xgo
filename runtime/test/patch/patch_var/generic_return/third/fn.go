//go:build go1.18
// +build go1.18

package third

type Tree[Q any] struct {
	Root Q
}

func MustBuild[Q any](root Q) *Tree[Q] {
	return &Tree[Q]{
		Root: root,
	}
}
