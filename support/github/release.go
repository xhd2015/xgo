package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/httputil"
)

func GetLatestReleaseURL(repo string) string {
	return fmt.Sprintf("https://github.com/%s/releases/latest", repo)
}

func GetLatestReleaseTag(ctx context.Context, latestURL string) (string, error) {
	// example location:
	//   https://github.com/xhd2015/xgo/releases/tag/v1.0.31
	location, err := httputil.Get302Location(ctx, latestURL)
	if err != nil {
		return "", err
	}
	location = strings.TrimSpace(location)
	const anchor = "/releases/tag/"
	idx := strings.LastIndex(location, anchor)
	var version string
	if idx >= 0 {
		version = location[idx+len(anchor):]
	}

	if version == "" {
		return "", fmt.Errorf("%s does match '*%sVERSION'", anchor, location)
	}
	return version, nil
}
