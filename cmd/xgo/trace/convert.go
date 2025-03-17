package trace

import (
	"time"

	"github.com/xhd2015/xgo/cmd/xgo/trace/render"
)

func convert(root *RootExport) *render.Stack {
	return &render.Stack{
		Begin:    root.Begin.Format(time.RFC3339),
		Children: convertStackEntries(root.Children),
	}
}

func convertStackEntries(list []*StackExport) []*render.StackEntry {
	if list == nil {
		return nil
	}
	entries := make([]*render.StackEntry, len(list))
	for i, child := range list {
		entries[i] = &render.StackEntry{
			FuncInfo: convertFuncInfo(child.FuncInfo),
			BeginNs:  child.Begin,
			EndNs:    child.End,
			Args:     child.Args,
			Results:  child.Results,
			Panic:    child.Panic,
			Error:    child.Error,
			Children: convertStackEntries(child.Children),
		}
	}
	return entries
}

func convertFuncInfo(info *FuncInfoExport) *render.FuncInfo {
	return &render.FuncInfo{
		Kind:     render.FuncKind(info.Kind),
		Pkg:      info.Pkg,
		Name:     info.IdentityName, // NOTE: use IdentityName as display name
		RecvType: info.RecvType,
		RecvPtr:  info.RecvPtr,

		Interface: info.Interface,
		Generic:   info.Generic,
		Closure:   info.Closure,
		Stdlib:    info.Stdlib,

		File: info.File,
		Line: info.Line,

		RecvName: info.RecvName,
		ArgNames: info.ArgNames,
		ResNames: info.ResNames,

		FirstArgCtx:   info.FirstArgCtx,
		LastResultErr: info.LastResultErr,
	}
}
