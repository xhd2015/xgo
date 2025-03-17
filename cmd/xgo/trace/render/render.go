package render

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func ReadStacks(file string) ([]*Stack, bool, error) {
	return readStacks(file, false)
}

func readStacks(file string, force bool) ([]*Stack, bool, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, false, err
	}

	var stacks []*Stack
	decoder := json.NewDecoder(reader)
	// decode first
	var stack Stack
	err = decoder.Decode(&stack)
	if err != nil {
		if err == io.EOF {
			return nil, true, nil
		}
		return nil, false, err
	}
	if !force && stack.Format != "stack" {
		return nil, false, nil
	}
	stacks = append(stacks, &stack)
	for {
		var stack Stack
		err = decoder.Decode(&stack)
		if err != nil {
			if err == io.EOF {
				break
			}
			stack = Stack{
				Begin: time.Now().Format(time.RFC3339),
				Children: []*StackEntry{
					{
						Error: fmt.Sprintf("parse stack: %v", err),
					},
				},
			}
		}
		stacks = append(stacks, &stack)
	}
	return stacks, true, nil
}

func RenderFile(file string, w io.Writer) error {
	stacks, _, err := readStacks(file, true)
	if err != nil {
		return err
	}
	return RenderStacks(stacks, file, w)
}

func RenderStacks(stacks []*Stack, file string, w io.Writer) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
			} else {
				err = fmt.Errorf("panic:%v", e)
			}
		}
	}()

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
	var tops []*StackEntry
	for _, stack := range stacks {
		top := &StackEntry{
			FuncInfo: &FuncInfo{
				Name: "<root>",
			},
			Children: stack.Children,
		}
		walk(top)
		tops = append(tops, top)
	}

	if len(tops) > 1 {
		for i, top := range tops {
			top.FuncInfo.Name = fmt.Sprintf("<root_%d>", i)
		}
	}

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
	for _, top := range tops {
		renderItem(h, top, traceIDMapping)
	}
	h("</ul>")
	h(`</div>`)

	vscode := vscodeIconSVG
	vscode = strings.Replace(vscode, `width="100"`, `width="14"`, 1)
	vscode = strings.Replace(vscode, `height="100"`, `height="14"`, 1)

	vscode = `<div id="vscode-icon" class="vscode-icon" onclick="onClickVscodeIcon(arguments[0])">` + vscode + `</div>`

	// Create copy icon with same replacements as vscode icon
	copyChecked := fmt.Sprintf(`<span class="copy-icon-action">%s</span><span class="copy-icon-checked">%s</span>`, copyIconSVG, checkedIconSVG)
	copy := `<div id="copy-icon" class="copy-icon" onclick="onClickCopyIcon(arguments[0])">` + copyChecked + `</div>`

	h(`<div class="detail">`)
	h(`<div id="detail-info">
	   <div class="label-value"> <label>Pkg:</label>	   <div id="detail-info-pkg"> </div> </div>
	   <div class="label-value"> <label>Func:</label>    <div id="detail-info-func"> </div> ` + copy + vscode + `</div>
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
	return nil
}
