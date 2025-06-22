//go:build go1.22
// +build go1.22

package tasker

import "context"

type Tasker[T any] struct {
	Client T
	F      func(context.Context, T) error
}

func (t Tasker[T]) Do(ctx context.Context) error {
	return t.F(ctx, t.Client)
}
