package patch

import (
	"context"
	"fmt"
)

func toErr(s string) error {
	return fmt.Errorf("err: %v", s)
}

func nilCtx(a int, ctx context.Context) {
	panic("nilCtx should be mocked")
}
