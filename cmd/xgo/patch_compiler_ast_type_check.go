package main

import (
	"os"

	"github.com/xhd2015/xgo/support/goinfo"
)

const convertXY = `
if xgoConv, ok := x.expr.(*syntax.XgoSimpleConvert); ok {
	var isConst bool
	switch y.expr.(type) {
	    case *syntax.XgoSimpleConvert,*syntax.BasicLit:
			isConst=true
	}
	if !isConst{
		t := y.typ

		callExpr := xgoConv.X.(*syntax.CallExpr)
		ct := callExpr.GetTypeInfo()
		ct.Type = t
		callExpr.SetTypeInfo(ct)

		name := callExpr.Fun.(*syntax.Name)
		nt := name.GetTypeInfo()
		nt.Type = t
		name.SetTypeInfo(nt)
		name.Value = t.String()

		xt := xgoConv.GetTypeInfo()
		xt.Type = t
		xgoConv.SetTypeInfo(xt)

		x.typ = t
	}
}else if xgoConv,ok := y.expr.(*syntax.XgoSimpleConvert);ok {
	var isConst bool
	switch x.expr.(type) {
	    case *syntax.XgoSimpleConvert,*syntax.BasicLit:
			isConst=true
	}
	if !isConst{
		t := x.typ
		callExpr := xgoConv.X.(*syntax.CallExpr)
		ct := callExpr.GetTypeInfo()
		ct.Type = t
		callExpr.SetTypeInfo(ct)

		name := callExpr.Fun.(*syntax.Name)
		nt := name.GetTypeInfo()
		nt.Type = t
		name.SetTypeInfo(nt)
		name.Value = t.String()

		xt := xgoConv.GetTypeInfo()
		xt.Type = t
		xgoConv.SetTypeInfo(xt)

		y.typ = t
	}
}
`

func getExprInternalPatch(mark string, rawCall string, checkGoVersion func(goVersion *goinfo.GoVersion) bool) *Patch {
	return &Patch{
		Mark:         mark,
		InsertIndex:  5,
		InsertBefore: true,
		Anchors: []string{
			`(check *Checker) exprInternal`,
			"\n",
			`default:`,
			`case *syntax.Operation:`,
			`case *syntax.KeyValueExpr:`,
			`default:`,
			"\n",
		},
		Content: `
case *syntax.XgoSimpleConvert:
kind := check.` + rawCall + `
x.expr = e
return kind
`,
		CheckGoVersion: checkGoVersion,
	}
}

var type2ExprPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "types2", "expr.go"},
	Patches: []*Patch{
		getExprInternalPatch("type2_check_xgo_simple_convert", `rawExpr(nil, x, e.X, nil, false)`, func(goVersion *goinfo.GoVersion) bool {
			return goVersion.Major > 1 || goVersion.Minor >= 21
		}),
		getExprInternalPatch("type2_check_xgo_simple_convert_no_target", `rawExpr(x, e.X, nil, false)`, func(goVersion *goinfo.GoVersion) bool {
			return goVersion.Major == 1 && goVersion.Minor < 21
		}),
		{
			Mark:        "type2_match_type_xgo_simple_convert",
			InsertIndex: 1,
			Anchors: []string{
				`func (check *Checker) matchTypes(x, y *operand) {`,
				"\n",
			},
			Content: convertXY,
			CheckGoVersion: func(goVersion *goinfo.GoVersion) bool {
				return goVersion.Major > 1 || goVersion.Minor >= 21
			},
		},
		{
			Mark:         "type2_binary_convert_type_xgo_simple_convert",
			InsertIndex:  2,
			InsertBefore: true,
			Anchors: []string{
				`func (check *Checker) binary(x *operand`,
				"\n",
				`mayConvert := func(x, y *operand) bool {`,
			},
			Content: `
			(func(x, y *operand){` + convertXY + `})(x,&y)
			`,
			CheckGoVersion: func(goVersion *goinfo.GoVersion) bool {
				return goVersion.Major == 1 && goVersion.Minor >= 20 && goVersion.Minor < 21
			},
		},
		{
			Mark:         "type2_binary_convert_type_xgo_simple_convert_can_mix",
			InsertIndex:  2,
			InsertBefore: true,
			Anchors: []string{
				`func (check *Checker) binary(x *operand`,
				"\n",
				`canMix := func(x, y *operand) bool {`,
			},
			Content: `
			(func(x, y *operand){` + convertXY + `})(x,&y)
			`,
			CheckGoVersion: func(goVersion *goinfo.GoVersion) bool {
				return goVersion.Major == 1 && goVersion.Minor < 20
			},
		},
		{
			Mark:        "type2_comparison_xgo_simple_convert",
			InsertIndex: 2,
			Anchors: []string{
				`func (check *Checker) comparison(x, y *operand`,
				"{",
				"\n",
			},
			Content: convertXY,
		},
	},
}

var type2AssignmentsPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "types2", "assignments.go"},
	Patches: []*Patch{
		{
			Mark:         "type2_assignment_rewrite_xgo_simple_convert",
			InsertIndex:  1,
			InsertBefore: true,
			Anchors: []string{
				`func (check *Checker) assignment(`,
				`switch x.mode {`,
			},
			Content: `
			if xgoConv, ok := x.expr.(*syntax.XgoSimpleConvert); ok {
				callExpr := xgoConv.X.(*syntax.CallExpr)
				funName := callExpr.Fun.(*syntax.Name)
				t := funName.GetTypeInfo()
				t.Type = T
				funName.SetTypeInfo(t)
				funName.Value = T.String()
		
				ct := callExpr.GetTypeInfo()
				ct.Type = T
				callExpr.SetTypeInfo(ct)
		
				x.expr = callExpr
				x.typ = T
				xt := xgoConv.GetTypeInfo()
				xt.Type = T
				xgoConv.SetTypeInfo(xt)
			}
			`,
		},
	},
}

var syntaxWalkPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "syntax", "walk.go"},
	Patches: []*Patch{
		{
			Mark:         "syntax_walk_xgo_simple_convert",
			InsertIndex:  4,
			InsertBefore: true,
			Anchors: []string{
				`func (w walker) node(n Node) {`,
				`case *RangeClause:`,
				`case *CaseClause:`,
				`case *CommClause:`,
				`default`,
			},
			Content: `
		case *XgoSimpleConvert:
			w.node(n.X)
			`,
		},
	},
}

var noderWriterPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "noder", "writer.go"},
	Patches: []*Patch{
		{
			Mark:         "noder_write_xgo_simple_convert",
			InsertIndex:  3,
			InsertBefore: true,
			Anchors: []string{
				`func (w *writer) expr(expr syntax.Expr) {`,
				`switch expr := expr.(type) {`,
				`case *syntax.Operation:`,
				`case *syntax.CallExpr:`,
			},
			Content: `
		case *syntax.XgoSimpleConvert:
			w.expr(expr.X)
			`,
		},
	},
}
var syntaxExtra = _FilePath{"src", "cmd", "compile", "internal", "syntax", "xgo_extra.go"}

const syntaxExtraPatch = `
package syntax

// helper:  convert anything to which
// the type is expected
type XgoSimpleConvert struct {
	X Expr
	expr
}
`

func patchCompilerAstTypeCheck(goroot string, goVersion *goinfo.GoVersion) error {
	// always generate xgo_extra file
	syntaxExtraFile := syntaxExtra.Join(goroot)
	err := os.WriteFile(syntaxExtraFile, []byte(syntaxExtraPatch), 0755)
	if err != nil {
		return err
	}
	if goVersion.Major == 1 && goVersion.Minor < 20 {
		// only go1.20 and above supports const mock
		return nil
	}
	err = type2ExprPatch.Apply(goroot, goVersion)
	if err != nil {
		return err
	}
	err = type2AssignmentsPatch.Apply(goroot, goVersion)
	if err != nil {
		return err
	}
	err = syntaxWalkPatch.Apply(goroot, goVersion)
	if err != nil {
		return err
	}
	err = noderWriterPatch.Apply(goroot, goVersion)
	if err != nil {
		return err
	}
	return nil
}
