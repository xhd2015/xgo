package downloadgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// List returns naked Go versions parsed from go.dev HTML.
func List(ctx context.Context, opts ListOptions) ([]string, error) {
	fetch := opts.FetchHTML
	if fetch == nil {
		fetch = defaultFetchHTML
	}
	html, err := fetch(ctx)
	if err != nil {
		return nil, err
	}
	return parseDownloadVersions(html), nil
}

func defaultFetchHTML(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", downloadListURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("statusCode=%d, resp=%s", resp.StatusCode, bodyData)
	}
	return string(bodyData), nil
}

func parseDownloadVersions(htmlContent string) []string {
	var goVersions []string
	// find all div like id="go1.22.1"
	lines := strings.Split(htmlContent, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "<div ") {
			continue
		}
		const idGo = `id="go`
		idx := strings.Index(line, idGo)
		if idx < 0 {
			continue
		}
		base := idx + len(idGo)
		qIdx := strings.Index(line[base:], `"`)
		if qIdx < 0 {
			continue
		}
		goVersion := line[base : base+qIdx]
		if goVersion == "" {
			continue
		}
		goVersions = append(goVersions, goVersion)
	}
	return goVersions
}
