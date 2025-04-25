package syntax

import (
	"cmd/compile/internal/syntax"
	"fmt"
)

func fillMissingArgNames(fn *syntax.FuncDecl) {
	if fn.Recv != nil {
		fillName(fn.Recv, "__xgo_recv_auto_filled")
	}
	for i, p := range fn.Type.ParamList {
		fillName(p, fmt.Sprintf("__xgo_arg_auto_filled_%d", i))
	}
}

func fillName(field *syntax.Field, namePrefix string) {
	if field.Name == nil {
		field.Name = syntax.NewName(field.Pos(), namePrefix)
		return
	}
	if field.Name.Value == "_" {
		field.Name.Value = namePrefix + "_blank"
		return
	}
}

// prevent all ident appeared in func type
// NOTE: the returned map may contain "_", ""
func getPresetNames(node syntax.Node) map[string]bool {
	preset := make(map[string]bool)
	syntax.Inspect(node, func(n syntax.Node) bool {
		if n == nil {
			return false
		}
		if idt, ok := n.(*syntax.Name); ok {
			preset[idt.Value] = true
		}
		return true
	})
	return preset
}

func isBlankName(name string) bool {
	return name == "" || name == "_"
}

func getFieldNames(fields []*syntax.Field) []string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, getFieldName(field))
	}
	return names
}
func getFieldName(f *syntax.Field) string {
	if f == nil || f.Name == nil {
		return ""
	}
	return f.Name.Value
}

// auto fill unnamed parameters
func fillNames(pos syntax.Pos, recv *syntax.Field, funcType *syntax.FuncType, preset map[string]bool) {
	if recv != nil {
		fillFieldNames(pos, []*syntax.Field{recv}, preset, "_x")
	}
	fillFieldNames(pos, funcType.ParamList, preset, "_a")
	fillFieldNames(pos, funcType.ResultList, preset, "_r")
}

type ISetPos interface {
	SetPos(p syntax.Pos)
}

func fillPos(pos syntax.Pos, node syntax.Node) {
	syntax.Inspect(node, func(n syntax.Node) bool {
		if n == nil {
			return false
		}
		n.(ISetPos).SetPos(pos)
		return true
	})
}
