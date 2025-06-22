//go:build go1.22
// +build go1.22

package mypackage_test

import (
	"context"
	"testing"

	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xhd2015/xgo/runtime/test/mock/mock_var/mock_var_generic_issue_357/mypackage"
)

func TestDo(t *testing.T) {
	mock.Patch(mypackage.DoTask, func(ctx context.Context) error {
		t.Log("patched DoTask")
		return nil
	})
	err := mypackage.DoTask(context.Background())
	if err != nil {
		t.Errorf("mypackage failed: %s", err)
	}
}
