package render

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed style.css
var styles string

//go:embed script.js
var script string

//go:embed vscode.svg
var vscodeIconSVG string

// Add copy and check icons
var copyIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
	<rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
	<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
</svg>`

var checkedIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
	<path d="M20 6L9 17l-5-5"></path>
</svg>`

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

func renderToolbar(h func(s string)) {
	h(fmt.Sprintf(`<div id="toolbar" class="toggle-all-on" onClick="onClickExpandAll(arguments[0])">%s</div>`, svgExpand))
}

func getTraceListID(id int64) string {
	return fmt.Sprintf("trace_list_%d", id)
}
