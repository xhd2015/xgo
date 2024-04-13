// debug test is a convenient package
// you can paste your minimal code your
// to focus only the problemtic part of
// failing code

package debug

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/xhd2015/xgo/runtime/core"
	"github.com/xhd2015/xgo/runtime/trap"
)

func TestTimeNowNestedLevel2AllowNested(t *testing.T) {
	i := 0
	trap.AddInterceptor(&trap.Interceptor{
		Pre: func(ctx context.Context, f *core.FuncInfo, args, result core.Object) (data interface{}, err error) {
			// t.Logf("%s.%s", f.Pkg, f.IdentityName)
			i++
			if i > 20 {
				os.Exit(1)
			}
			getTime2()
			return
		},
	})
	time.Now()
}

func getTime2() time.Time {
	return getTime3()
}

func getTime3() time.Time {
	return time.Now()
}
