package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/trace/render/stack_model"
)

const allowPkgName = false

func marshalStackWithoutChildren(stack *stack_model.StackEntry) ([]byte, error) {
	s := stack.Children
	stack.Children = nil
	defer func() {
		stack.Children = s
	}()
	return json.Marshal(stack)
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
