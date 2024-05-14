package httputil

import (
	"context"
	"fmt"
	"net/http"
)

func Get302Location(ctx context.Context, url string) (string, error) {
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 302 {
		return "", fmt.Errorf("expect 302 from %s", url)
	}

	loc, err := resp.Location()
	if err != nil {
		return "", err
	}
	return loc.Path, nil
}
