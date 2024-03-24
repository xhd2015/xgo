package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/xgo/support/cmd"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "requires file\n")
		os.Exit(1)
	}
	file := args[0]
	serveFile(file)
}

func serveFile(file string) {
	server := http.NewServeMux()
	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				stack := debug.Stack()
				io.WriteString(w, fmt.Sprintf("<pre>panic: %v\n%s</pre>", e, stack))
			}
		}()
		record, err := parseRecord(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, fmt.Sprintf("%v", err))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		renderRecordHTML(record, file, w)
	})
	server.HandleFunc("/openVscodeFile", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		file := q.Get("file")
		if file == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "no file\n")
			return
		}
		_, err := os.Stat(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, err.Error())
			return
		}
		line := q.Get("line")

		fileLine := file
		if line != "" {
			fileLine = file + ":" + line
		}
		output, err := exec.Command("code", "--goto", fileLine).Output()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if exitErr, ok := err.(*exec.ExitError); ok {
				w.Write(exitErr.Stderr)
			} else {
				io.WriteString(w, err.Error())
			}
			return
		}
		w.Write(output)
	})
	port := 7070
	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("Server listen at %s\n", url)

	go func() {
		time.Sleep(500 * time.Millisecond)
		openCmd := "open"
		if runtime.GOOS == "windows" {
			openCmd = "explorer"
		}
		cmd.Run(openCmd, url)
	}()

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), server)
	if err != nil {
		panic(err)
	}
}

//go:embed style.css
var styles string

//go:embed script.js
var script string

//go:embed vscode.svg
var vscodeIconSVG string

func parseRecord(file string) (*RootExport, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var root *RootExport
	err = json.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func renderRecordHTML(root *RootExport, file string, w io.Writer) {
	h := func(s string) {
		_, err := io.WriteString(w, s)
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "\n")
		if err != nil {
			panic(err)
		}
	}
	h(`<!DOCTYPE html>
	<html lang="en" style="height: 100%;">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Trace of ` + file + `</title>
	</head>
	<body style="height: 100%;">
	`,
	)

	h(`<div style="height: 100%;">`)
	h(`<style>`)
	h(styles)
	h(`</style>`)

	top := &StackExport{
		FuncInfo: &FuncInfoExport{
			IdentityName: "<root>",
		},
		Children: root.Children,
	}

	h("<script>")
	h("window.onload = function(){")
	h(" const traces = {}")
	h(" const ids = []")
	traceIDMapping := make(map[*StackExport]int64)
	nextID := int64(1)
	var walk func(stack *StackExport)
	walk = func(stack *StackExport) {
		id := nextID
		nextID++
		traceIDMapping[stack] = id

		stackData, err := marshalStackWithoutChildren(stack)
		if err != nil {
			stackData = []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
		}
		h(fmt.Sprintf(` traces["%d"] = %s`, id, stackData))
		h(fmt.Sprintf(` ids.push("%d")`, id))
		for _, child := range stack.Children {
			walk(child)
		}
	}
	walk(top)

	h(script)
	h("}")
	h("</script>")

	h(`<div class="root">`)

	h(`<div class="trace-list-root">`)
	h(`<div>`)
	renderToolbar(h)
	h(`</div>`)
	// h(fmt.Sprintf(`<ul id="%s" class="trace-list">`, getTraceListID(traceIDMapping[top])))
	h(`<ul class="trace-list">`)
	add(h, top, traceIDMapping)
	h("</ul>")
	h(`</div>`)

	vscode := vscodeIconSVG
	vscode = strings.Replace(vscode, `width="100"`, `width="14"`, 1)
	vscode = strings.Replace(vscode, `height="100"`, `height="14"`, 1)

	vscode = `<div id="vscode-icon" class="vscode-icon" onclick="onClickVscodeIcon(arguments[0])">` + vscode + `</div>`

	h(`<div class="detail">`)
	h(`<div id="detail-info">
	   <div class="label-value"> <label>Pkg:</label>	   <div id="detail-info-pkg"> </div> </div>
	   <div class="label-value"> <label>Func:</label>    <div id="detail-info-func"> </div> ` + vscode + `</div>
	</div>`)
	h(`<label>Request</label>`)
	h(`<textarea id="detail-request"  placeholder="request..."></textarea>`)
	h(`<label>Response</label>`)
	h(`<textarea id="detail-response" placeholder="response..."></textarea>`)
	h("</div>")

	h("</div>")

	h("</div>")
	h(`</body>
	</html>`)
}
func renderToolbar(h func(s string)) {
	h(fmt.Sprintf(`<div id="toolbar" class="toggle-all-on" onClick="onClickExpandAll(arguments[0])">%s</div>`, svgExpand))
}

func getTraceListID(id int64) string {
	return fmt.Sprintf("trace_list_%d", id)
}

const svgIconDown = `<svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708z"></path></svg>`

const svgIconRight = `<svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M4.646 1.646a.5.5 0 0 1 .708 0l6 6a.5.5 0 0 1 0 .708l-6 6a.5.5 0 0 1-.708-.708L10.293 8 4.646 2.354a.5.5 0 0 1 0-.708z"></path></svg>`

const svgExpand = `<svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 16 16" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><path d="M9 9H4v1h5V9z"></path><path fill-rule="evenodd" clip-rule="evenodd" d="M5 3l1-1h7l1 1v7l-1 1h-2v2l-1 1H3l-1-1V6l1-1h2V3zm1 2h4l1 1v4h2V3H6v2zm4 1H3v7h7V6z"></path></svg>`

func makeSvg(svg string, extra string) string {
	if extra == "" {
		return svg
	}
	mark := "<svg "
	idx := strings.Index(svg, "<svg ")
	if idx < 0 {
		panic(fmt.Errorf("<svg not found"))
	}
	return svg[:idx+len(mark)] + extra + " " + svg[idx+len(mark):]
}

func marshalStackWithoutChildren(stack *StackExport) ([]byte, error) {
	s := stack.Children
	stack.Children = nil
	defer func() {
		stack.Children = s
	}()
	return json.Marshal(stack)
}

const allowPkgName = false

func add(h func(string), stack *StackExport, traceIDMapping map[*StackExport]int64) {
	var name string
	if stack.FuncInfo != nil {
		name = stack.FuncInfo.IdentityName
		if stack.FuncInfo.Pkg != "" && allowPkgName {
			name = lastPart(stack.FuncInfo.Pkg) + "." + name
		}
	}
	if name == "" {
		name = "<unknown>"
	}
	id := traceIDMapping[stack]

	var indicator string

	if len(stack.Children) > 0 {
		// NOTE: onclick on svg does not work, must wrap it with div
		makeIcon := func(status string, down string, right string) string {
			svgDown := makeSvg(down, `class="toggle-icon-down"`)
			svgRight := makeSvg(right, `class="toggle-icon-right"`)
			toggleID := fmt.Sprintf("toggle_%d", id)
			return fmt.Sprintf(`<div  id="%s" class="toggle %s" onclick="onClickToggle(arguments[0],'%d')">%s%s</div>`, toggleID, status, id, svgDown, svgRight)
		}
		indicator = makeIcon("down", svgIconDown, svgIconRight)
	}
	headClass := "head-block"
	if stack.Panic {
		headClass = headClass + " panic"
	}
	if stack.Error != "" {
		headClass = headClass + " error"
	}

	h(fmt.Sprintf(`<div class="head">
	%s
	<div class="head-info" id="head_%d" onclick="onClickHead('%d')">
		<div class="%s"></div>
		<span class="head-name">%s</span>
		<span class="head-cost">%s</span>
	</div>
	</div>
	`,
		indicator,
		id, id,
		headClass,
		html.EscapeString(name),
		formatCost(stack.Begin, stack.End),
	))

	if len(stack.Children) == 0 {
		return
	}
	h(fmt.Sprintf(`<ul id="%s" class="trace-sub-list">`, getTraceListID(id)))
	for _, child := range stack.Children {
		h("<li>")
		add(h, child, traceIDMapping)
		h("</li>")
	}
	h("</ul>")
}

func lastPart(pkg string) string {
	idx := strings.LastIndex(pkg, "/")
	if idx < 0 {
		return pkg
	}
	return pkg[idx+1:]
}

func formatCost(begin int64, end int64) string {
	if end == 0 && begin == 0 {
		return ""
	}
	cost := end - begin
	var sign string
	if cost < 0 {
		sign = "-"
		cost = -cost
	}
	type unit struct {
		name      string
		scaleLast int
	}
	var units = []unit{
		{"ns", 1},
		{"μs", 1000},
		{"ms", 1000},
		{"s", 1000},
		{"m", 60},
		{"h", 60},
		{"d", 24},
	}
	unitName := units[0].name
	f := float64(cost)
	for i := 1; i < len(units); i++ {
		x := float64(units[i].scaleLast)
		if f < x {
			break
		}
		f = f / x
		unitName = units[i].name
	}

	return fmt.Sprintf("%s%d%s", sign, int(f), unitName)
}
