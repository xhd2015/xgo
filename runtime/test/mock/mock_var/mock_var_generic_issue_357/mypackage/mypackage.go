//go:build go1.22
// +build go1.22

package mypackage

import (
	"context"
	"fmt"

	"github.com/xhd2015/xgo/runtime/test/mock/mock_var/mock_var_generic_issue_357/mypackage/myclient"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_var/mock_var_generic_issue_357/tasker"
)

var MyTasker = tasker.Tasker[myclient.Client]{
	Client: myclient.Client{URL: "https://example.com"},
	F: func(ctx context.Context, client myclient.Client) error {
		fmt.Printf("task called: %s", client.URL)
		return nil
	},
}

func DoTask(ctx context.Context) error {
	return MyTasker.Do(ctx)
}
