package git

import (
	"fmt"
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git/fetch"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/go-inspect/sh"
)

type FetchOptions struct {
	Timeout time.Duration
	Depth   int
}

func FetchAll(dir string, opts *FetchOptions) error {
	var timeout time.Duration
	var depth int
	if opts != nil {
		timeout = opts.Timeout
		depth = opts.Depth
	}
	fetchArgs := fetch.FormatFetch("", &fetch.Options{
		AllTags: true,
		Depth:   depth,
	})

	// "url=$(git remote get-url origin)",
	// "if [[ $url = *'/shopee/seamoney-investment/investment-be/txn/service'* ]];then exit 0;fi",
	_, err := RunCommandsWithOptions(dir, sh.RunBashOptions{
		Timeout: timeout,
	}, "git "+sh.Quotes(fetchArgs...))

	return err
}

func FetchSingle(dir string, origin string, ref string, opts *FetchOptions) error {
	if origin == "" {
		return fmt.Errorf("requires origin")
	}
	if ref == "" {
		return fmt.Errorf("requires ref")
	}
	if ref == COMMIT_WORKING {
		return nil
	}
	var timeout time.Duration
	var depth int
	if opts != nil {
		timeout = opts.Timeout
		depth = opts.Depth
	}

	fetchArgs := fetch.FormatFetch(origin, &fetch.Options{
		Branch:  ref,
		AllTags: true,
		Depth:   depth,
	})

	// "url=$(git remote get-url origin)",
	// "if [[ $url = *'/shopee/seamoney-investment/investment-be/txn/service'* ]];then exit 0;fi",
	_, err := RunCommandsWithOptions(dir, sh.RunBashOptions{
		Timeout: timeout,
	}, "git "+sh.Quotes(fetchArgs...))

	return err
}
