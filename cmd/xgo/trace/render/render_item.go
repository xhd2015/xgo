package render

import (
	"fmt"
	"html"
)

func renderItem(h func(string), stack *StackEntry, traceIDMapping map[*StackEntry]int64) {
	var name string
	if stack.FuncInfo != nil {
		name = stack.FuncInfo.Name
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
		formatCost(stack.BeginNs, stack.EndNs),
	))

	if len(stack.Children) == 0 {
		return
	}
	h(fmt.Sprintf(`<ul id="%s" class="trace-sub-list">`, getTraceListID(id)))
	for _, child := range stack.Children {
		h("<li>")
		renderItem(h, child, traceIDMapping)
		h("</li>")
	}
	h("</ul>")
}
