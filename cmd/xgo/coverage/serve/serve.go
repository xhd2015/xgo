package serve

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/gitops/git"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/load/loadcov"
	"github.com/xhd2015/xgo/cmd/xgo/internal/vendir/github.com/xhd2015/lines-annotation/model/coverage"
	"github.com/xhd2015/xgo/support/netutil"
)

//go:embed index.html
var indexHTML string

const urlPlaceholder = "http://localhost:8000"

func getHost(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// RouteServer install these endpoints:
// /                 ->    index.html
// /fileAnnotations  ->    dynamic coverage
// /fileDetail       ->    get file content
// /diffFileDetail   ->    get diff file content
func RouteServer(server *http.ServeMux, prefix string, getPort func() int, opts loadcov.LoadAllOptions) {
	dir := opts.Dir
	diffBase := opts.DiffBase
	server.HandleFunc(prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := strings.ReplaceAll(indexHTML, urlPlaceholder, getHost(getPort())+prefix)
		w.Write([]byte(html))
	})
	server.HandleFunc(prefix+"/fileAnnotations", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			log.Printf("start coverage")
			defer func() {
				log.Printf("end coverage")
			}()
			onlyChanged := true
			full := r.URL.Query().Get("full")
			if full == "true" || (full == "" && r.URL.Query().Has("full")) {
				onlyChanged = false
			}

			cloneOpts := opts
			cloneOpts.OnlyChangedFiles = onlyChanged
			project, err := loadcov.LoadAll(cloneOpts)
			if err != nil {
				return nil, err
			}
			// NOTE: return files instead of project
			return project.Files, nil
		})
	})
	server.HandleFunc(prefix+"/summary", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			log.Printf("start summary")
			defer func() {
				log.Printf("end summary")
			}()
			project, err := loadcov.LoadAll(opts)
			if err != nil {
				return nil, err
			}
			type LinkDetail struct {
				*coverage.Detail
				Link string `json:"link"`
			}
			var detail *coverage.Detail
			summary := loadcov.ComputeCoverageSummary(project)
			base := summary[""]
			if base != nil {
				var incremental *coverage.Item
				if base.Incremental != nil {
					incremental = base.Incremental
				}
				if incremental == nil {
					// typo
					incremental = base.Incrimental
				}
				detail = &coverage.Detail{
					Total:       base.Total,
					Incremental: incremental,
				}
			}
			return &LinkDetail{
				Detail: detail,
				Link:   prefix + "/",
			}, nil
		})
	})

	server.HandleFunc(prefix+"/fileDetail", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			return handleGetFile(r, dir, opts.Ref)
		})
	})

	server.HandleFunc(prefix+"/diffFileDetail", func(w http.ResponseWriter, r *http.Request) {
		netutil.SetCORSHeaders(w)
		netutil.HandleJSON(w, r, func(ctx context.Context, r *http.Request) (interface{}, error) {
			return handleGetFile(r, dir, diffBase)
		})
	})
}

type GetFileResp struct {
	Exists  bool   `json:"exists"`
	Content string `json:"content"`
}

func handleGetFile(r *http.Request, dir string, commitID string) (*GetFileResp, error) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		return nil, nil
	}

	ok, content, err := git.CatFile(dir, commitID, filename)
	if err != nil {
		return nil, err
	}
	return &GetFileResp{
		Exists:  ok,
		Content: content,
	}, nil
}
