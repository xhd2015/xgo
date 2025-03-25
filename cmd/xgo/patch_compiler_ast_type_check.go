package main

import (
	"os"

	"github.com/xhd2015/xgo/support/goinfo"
)

const convertXY = `
if xgoConv, ok := x.expr.(*syntax.XgoSimpleConvert); ok {
	var isConst bool
	if isUntyped(y.typ) {
		isConst = true
	} else {
		switch y.expr.(type) {
		case *syntax.XgoSimpleConvert, *syntax.BasicLit:
			isConst = true
		}
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
	if isUntyped(x.typ) {
		isConst = true
	} else {
		switch x.expr.(type) {
		case *syntax.XgoSimpleConvert, *syntax.BasicLit:
			isConst = true
		}
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
		Mark:           mark,
		InsertIndex:    5,
		UpdatePosition: true,
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
			Mark:           "type2_binary_convert_type_xgo_simple_convert",
			InsertIndex:    2,
			UpdatePosition: true,
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
			Mark:           "type2_binary_convert_type_xgo_simple_convert_can_mix",
			InsertIndex:    2,
			UpdatePosition: true,
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
			Mark:           "type2_assignment_rewrite_xgo_simple_convert",
			InsertIndex:    1,
			UpdatePosition: true,
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
			Mark:           "syntax_walk_xgo_simple_convert",
			InsertIndex:    4,
			UpdatePosition: true,
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

var syntaxParserPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "syntax", "parser.go"},
	Patches: []*Patch{
		// {
		// 	// import
		// 	Mark:        "syntax_parser_import_xgo_syntax",
		// 	InsertIndex: 2,
		// 	Anchors: []string{
		// 		`package syntax`,
		// 		`import (`,
		// 		"\n",
		// 	},
		// 	Content: `xgo_syntax "cmd/compile/internal/xgo_rewrite_internal/patch/syntax"`,
		// },
		{
			// NOTE: dependency injection
			Mark:           "syntax_parser_record_comment_declare",
			InsertIndex:    0,
			UpdatePosition: true,
			Anchors: []string{
				`func (p *parser) init(file *PosBase,`,
			},
			Content: `var RecordComments func(file *PosBase, line, col uint, comment string)`,
		},
		{
			Mark:        "syntax_parser_record_comments",
			InsertIndex: 6,
			Anchors: []string{
				`func (p *parser) init(file *PosBase,`,
				`p.scanner.init(`,
				`func(line`,
				`text`,
				`:=`,
				`commentText(msg)`,
				"\n",
			},
			Content: `
			if RecordComments != nil {
				RecordComments(file,line,col,text)
			}
			`,
		},
	},
}

var noderWriterPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "noder", "writer.go"},
	Patches: []*Patch{
		{
			Mark:           "noder_write_xgo_simple_convert",
			InsertIndex:    3,
			UpdatePosition: true,
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

var noderExprPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "noder", "expr.go"},
	Patches: []*Patch{
		{
			Mark:           "noder_expr_const_expr_op_xgo_simple_convert",
			InsertIndex:    1,
			UpdatePosition: true,
			Anchors: []string{
				`func constExprOp(expr syntax.Expr) ir.Op {`,
				`case *syntax.BasicLit:`,
			},
			Content: `
		case *syntax.XgoSimpleConvert:
			return constExprOp(expr.X)
			`,
			CheckGoVersion: func(goVersion *goinfo.GoVersion) bool {
				return goVersion.Major == 1 && goVersion.Minor <= 21
			},
		},
	},
}

// /Users/xhd2015/.xgo/go-instrument-dev/go1.22.2_Us_xh_Pr_xh_xg_go_go_e9d4e1e8/go1.22.2/src/cmd/compile/internal/syntax/printer.go
var syntaxPrinterPatch = &FilePatch{
	FilePath: _FilePath{"src", "cmd", "compile", "internal", "syntax", "printer.go"},
	Patches: []*Patch{
		{
			Mark:           "noder_syntax_print_xgo_simple_convert",
			InsertIndex:    1,
			UpdatePosition: true,
			Anchors: []string{
				`func (p *printer) printRawNode(n Node) {`,
				`case *BasicLit:`,
			},
			Content: `
		case *XgoSimpleConvert:
			p.printRawNode(n.X)
			`,
			CheckGoVersion: func(goVersion *goinfo.GoVersion) bool {
				return goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor <= goinfo.GO_VERSION_23
			},
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
	syntaxExtraFile := syntaxExtra.JoinPrefix(goroot)
	err := os.WriteFile(syntaxExtraFile, []byte(syntaxExtraPatch), 0755)
	if err != nil {
		return err
	}
	// var mock: comments
	if false {
		// this does not work, because comments are turned off
		err = syntaxParserPatch.Apply(goroot, goVersion)
		if err != nil {
			return err
		}
	}
	if goVersion.Major > 1 || goVersion.Minor >= 20 {
		// only go1.20 and above supports const mock
		err := patchCompilerForConstTrap(goroot, goVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

func patchCompilerForConstTrap(goroot string, goVersion *goinfo.GoVersion) error {
	err := type2ExprPatch.Apply(goroot, goVersion)
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
	if goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor <= goinfo.GO_VERSION_21 {
		err = noderExprPatch.Apply(goroot, goVersion)
		if err != nil {
			return err
		}
	}
	if goVersion.Major == goinfo.GO_MAJOR_1 && goVersion.Minor <= goinfo.GO_VERSION_23 {
		err = syntaxPrinterPatch.Apply(goroot, goVersion)
		if err != nil {
			return err
		}
	}
	return nil
}
