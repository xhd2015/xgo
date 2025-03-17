package render

import (
	"fmt"
	"io"
	"strings"
)

func RenderHTML(root *Stack, file string, w io.Writer) {
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

	top := &StackEntry{
		FuncInfo: &FuncInfo{
			Name: "<root>",
		},
		Children: root.Children,
	}

	h("<script>")
	h("window.onload = function(){")
	h(" const traces = {}")
	h(" const ids = []")
	traceIDMapping := make(map[*StackEntry]int64)
	nextID := int64(1)
	var walk func(stack *StackEntry)
	walk = func(stack *StackEntry) {
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
	renderItem(h, top, traceIDMapping)
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
