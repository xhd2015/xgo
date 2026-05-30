package interceptor_ctx

import (
	"context"
	"time"
)

type CustomContext struct {
	innerCtx context.Context
}

func (c *CustomContext) Deadline() (deadline time.Time, ok bool) {
	if c.innerCtx == nil {
		return
	}
	return c.innerCtx.Deadline()
}

func (c *CustomContext) Done() <-chan struct{} {
	if c.innerCtx == nil {
		return nil
	}
	return c.innerCtx.Done()
}

func (c *CustomContext) Err() error {
	if c.innerCtx == nil {
		return nil
	}
	return c.innerCtx.Err()
}

func (c *CustomContext) Value(key interface{}) any {
	if c.innerCtx == nil {
		return nil
	}
	return c.innerCtx.Value(key)
}

func NewCustomContext() *CustomContext {
	return &CustomContext{
		innerCtx: context.Background(),
	}
}
