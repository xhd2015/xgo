package code

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"sort"
	"strings"
)

type CleanOpts struct {
	ShouldFormat    func(n ast.Node) bool
	Log             bool
	LogIndent       int  // spaces between lines, default 1 space
	DisallowUnknown bool // disallow unknown
}

func Fclean(w io.Writer, n ast.Node, opts CleanOpts) {
	f := &formatter{
		w:    w,
		opts: opts,
	}
	f.cleanCode(n)
}

func FcleanList(w io.Writer, lister func(func(n ast.Node, last bool)), join string, opts CleanOpts) {
	f := &formatter{
		w:    w,
		opts: opts,
	}
	f.cleanList(lister, join)
}

func Clean(n ast.Node, opts CleanOpts) string {
	var b strings.Builder
	Fclean(&b, n, opts)
	return b.String()
}

// `last` is used to help to decide whether adding the joint symbol or not.
func CleanList(lister func(func(n ast.Node, last bool)), join string, opts CleanOpts) string {
	var b strings.Builder
	FcleanList(&b, lister, join, opts)
	return b.String()
}

type formatter struct {
	w    io.Writer
	opts CleanOpts

	level int
}

func logNode(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.FuncDecl:
		return n.Name.Name
	case *ast.Field:
		names := make([]string, 0, len(n.Names))
		for _, n := range n.Names {
			names = append(names, n.Name)
		}
		return strings.Join(names, ", ")
	case *ast.TypeSpec:
		return n.Name.Name
	}
	return ""
}
func (c *formatter) cleanCode(n ast.Node) {
	if n == nil {
		panic(fmt.Errorf("cleanCode encountered nil"))
	}
	if c.opts.ShouldFormat != nil && !c.opts.ShouldFormat(n) {
		return
	}
	if c.opts.Log {
		s := logNode(n)
		if s != "" {
			s = "[" + s + "]"
		}
		indent := c.opts.LogIndent
		if indent <= 0 {
			indent = 1
		}
		fmt.Printf("%s%T%s\n", strings.Repeat(" ", c.level*indent), n, s)
	}
	c.level++
	switch n := n.(type) {
	case *ast.Ident:
		c.add(n.Name)
	case *ast.SelectorExpr:
		c.cleanCode(n.X)
		c.add(".", n.Sel.Name)
	case *ast.BadExpr:
		c.add("bad")
	case *ast.CallExpr:
		c.cleanCode(n.Fun)
		c.add("(")
		c.cleanExprs(n.Args, ",")
		if n.Ellipsis.IsValid() {
			c.add("...")
		}
		c.add(")")
	case *ast.FieldList:
		c.cleanFields(n.List, ",")
	case *ast.Field:
		c.cleanIdents(n.Names, ",")
		if len(n.Names) > 0 {
			c.add(" ")
		}
		c.cleanCode(n.Type)
		if n.Tag != nil {
			c.add(" `")
			c.cleanCode(n.Tag)
			c.add("`")
		}
		// check Spec
	case *ast.TypeSpec:
		c.add("type ", n.Name.Name)
		c.handleTypeParamsForTypeSpec(n)
		if n.Assign.IsValid() {
			c.add("=")
		} else {
			c.add(" ")
		}
		c.cleanCode(n.Type)
	case *ast.ValueSpec:
		c.add("var ")
		c.cleanIdents(n.Names, ",")
		if len(n.Values) > 0 {
			c.add("=")
			c.cleanExprs(n.Values, ",")
		}
		// check Decl
	case *ast.FuncDecl:
		c.add("func ")
		if n.Recv != nil {
			c.add("(")
			c.cleanFields(n.Recv.List, ",")
			c.add(")")
		}
		c.add(n.Name.Name)
		oldFunc := n.Type.Func
		n.Type.Func = token.NoPos
		c.cleanCode(n.Type)
		n.Type.Func = oldFunc
		c.cleanCode(n.Body)
	case *ast.GenDecl:
		if n.Tok == token.IMPORT {
			// processed earlier
			return
		}
		// if n.TokPos.IsValid() {
		// 	c.add( n.Tok.String(), " ")
		// }
		if n.Lparen.IsValid() {
			c.add("(")
		}
		c.cleanSpecs(n.Specs, "\n")
		if n.Rparen.IsValid() {
			c.add(")")
		}
		// check Stmts
	case *ast.BlockStmt:
		if n.Lbrace.IsValid() {
			c.add("{")
		}
		c.cleanStmts(n.List, ";")
		if n.Rbrace.IsValid() {
			c.add("}")
		}
	case *ast.DeferStmt:
		c.add("defer ")
		c.cleanCode(n.Call)
	case *ast.GoStmt:
		c.add("go ")
		c.cleanCode(n.Call)
	case *ast.ReturnStmt:
		c.add("return")
		if n.Results != nil {
			c.add(" ")
			c.cleanExprs(n.Results, ",")
		}
	case *ast.IfStmt:
		c.add("if ")
		c.cleanCode(n.Cond)
		c.cleanCode(n.Body)
		if n.Else != nil {
			c.add("else ")
			c.cleanCode(n.Else)
		}
	case *ast.ForStmt:
		c.add("for ")
		if n.Init != nil {
			c.cleanCode(n.Init)
		}
		c.add(";")
		if n.Cond != nil {
			c.cleanCode(n.Cond)
		}
		c.add(";")
		if n.Post != nil {
			c.cleanCode(n.Post)
		}
		c.cleanCode(n.Body)
	case *ast.RangeStmt:
		c.add("for ")
		if n.Key != nil {
			c.cleanCode(n.Key)
			if n.Value != nil {
				c.add(",")
				c.cleanCode(n.Value)
			}
			c.add(n.Tok.String())
			c.add("range ")
		}
		c.cleanCode(n.X)
		c.cleanCode(n.Body)
	case *ast.BranchStmt: // break,goto,continue
		c.add(n.Tok.String())
		if n.Label != nil {
			c.add(" ")
			c.add(n.Label.Name)
		}
	case *ast.SwitchStmt:
		c.add("switch ")
		if n.Init != nil {
			c.cleanCode(n.Init)
			c.add(";")
		}
		if n.Tag != nil {
			c.cleanCode(n.Tag)
		}
		c.cleanCode(n.Body)
	case *ast.SelectStmt:
		c.add("select ")
		c.cleanCode(n.Body)
	case *ast.TypeSwitchStmt:
		c.add("switch ")
		if n.Init != nil {
			c.cleanCode(n.Init)
			c.add(";")
		}
		if n.Assign != nil {
			c.cleanCode(n.Assign)
		}
		c.cleanCode(n.Body)
	case *ast.CaseClause:
		if n.List != nil {
			c.add("case ")
			c.cleanExprs(n.List, ",")
			c.add(":")
		} else {
			c.add("default:")
		}
		c.cleanStmts(n.Body, ";")
	case *ast.DeclStmt:
		c.cleanCode(n.Decl)
	case *ast.IncDecStmt: // ++, --
		c.cleanCode(n.X)
		c.add(n.Tok.String())
		// expr
	case *ast.ExprStmt:
		c.cleanCode(n.X)
	case *ast.AssignStmt:
		// n.Tok
		c.cleanExprs(n.Lhs, ",")
		c.add(n.Tok.String())
		c.cleanExprs(n.Rhs, ",")
		// check Expr
	case *ast.SendStmt:
		c.cleanCode(n.Chan)
		c.add("<-")
		c.cleanCode(n.Value)
	case *ast.BinaryExpr:
		c.cleanCode(n.X)
		c.add(n.Op.String())
		c.cleanCode(n.Y)
	case *ast.StarExpr:
		c.add("*")
		c.cleanCode(n.X)
	case *ast.UnaryExpr: // !done
		c.add(n.Op.String())
		c.cleanCode(n.X)
	case *ast.KeyValueExpr:
		c.cleanCode(n.Key)
		c.add(":")
		c.cleanCode(n.Value)
		c.add(",")
	case *ast.Ellipsis:
		c.add("...")
		if n.Elt != nil {
			c.cleanCode(n.Elt)
		}
	case *ast.TypeAssertExpr:
		c.cleanCode(n.X)
		c.add(".")
		c.add("(")
		if n.Type != nil {
			c.cleanCode(n.Type)
		} else {
			c.add("type")
		}
		c.add(")")
		// check Type
	case *ast.FuncType:
		// when from *FuncDecl, the 'func' and `name`
		// is prepended already
		if n.Func.IsValid() {
			c.add("func")
		}
		c.handleTypeParams(n)
		c.add("(")
		c.cleanFields(n.Params.List, ",")
		c.add(")")
		if n.Results != nil {
			if len(n.Results.List) > 1 {
				c.add("(")
			}
			c.cleanFields(n.Results.List, ",")
			if len(n.Results.List) > 1 {
				c.add(")")
			}
		}
	case *ast.StructType:
		if n.Struct.IsValid() {
			c.add("struct")
		}
		c.add("{")
		c.cleanFields(n.Fields.List, ",")
		c.add("}")
	case *ast.ArrayType:
		c.add("[")
		if n.Len != nil {
			c.cleanCode(n.Len)
		}
		c.add("]")
		c.cleanCode(n.Elt)
	case *ast.MapType:
		c.add("map[")
		c.cleanCode(n.Key)
		c.add("]")
		c.cleanCode(n.Value)
	case *ast.InterfaceType:
		c.add("interface")
		c.cleanCode(n.Methods)
	case *ast.ChanType:
		if n.Arrow == token.NoPos {
			c.add("chan ")
		} else if n.Arrow < n.Begin {
			c.add("<-chan ")
		} else {
			c.add("chan<- ")
		}
		c.cleanCode(n.Value)
	// check Literal
	case *ast.BasicLit:
		c.add(n.Value)
	case *ast.FuncLit:
		c.cleanCode(n.Type)
		c.cleanCode(n.Body)
	case *ast.CompositeLit: // []int{1,2,3}
		if n.Type != nil {
			c.cleanCode(n.Type)
		}
		c.add("{")
		c.cleanExprs(n.Elts, ",")
		c.add("}")
	case *ast.ParenExpr: // if (a>20)
		if n.Lparen.IsValid() {
			c.add("(")

		}
		c.cleanCode(n.X)
		if n.Rparen.IsValid() {
			c.add(")")
		}
	case *ast.SliceExpr:
		c.cleanCode(n.X)
		c.add("[")
		if n.Low != nil {
			c.cleanCode(n.Low)
		}
		c.add(":")

		if n.High != nil {
			c.cleanCode(n.High)
		}
		if n.Slice3 {
			c.add(":")
			if n.Max != nil {
				c.cleanCode(n.Max)
			}
		}
		c.add("]")
	case *ast.IndexExpr:
		c.cleanCode(n.X)
		c.add("[")
		c.cleanCode(n.Index)
		c.add("]")

	case *ast.CommClause, *ast.Comment, *ast.CommentGroup:
		// ignore any comment
		// return ""
	case *ast.File:
		c.add(fmt.Sprintf("package %s\n", n.Name.Name))
		if len(n.Imports) > 0 {
			// imps := make([]*ast.ImportSpec, len(n.Imports))
			// copy(imps, n.Imports)
			imps := append([]*ast.ImportSpec(nil), n.Imports...)
			sort.Slice(imps, func(i, j int) bool {
				return imps[i].Path.Value < imps[j].Path.Value
			})
			c.add("import (\n")
			for _, imp := range imps {
				if imp.Name != nil {
					c.add(imp.Name.Name, " ")
				}
				c.add("    ", imp.Path.Value)
				c.add("\n")
			}
			c.add(")\n")
		}
		for i, decl := range n.Decls {
			c.cleanCode(decl)
			if i < len(n.Decls)-1 {
				c.add("\n")
			}
		}
	default:
		if !c.handleFallback(n) {
			if c.opts.DisallowUnknown {
				panic(fmt.Errorf("unhandled %T", n))
			}
			c.add(fmt.Sprintf("TODO: %T", n))
		}
	}
	c.level--
}

func (c *formatter) cleanList(lister func(func(n ast.Node, last bool)), join string) {
	lister(func(n ast.Node, last bool) {
		c.cleanCode(n)
		if !last {
			fmt.Fprint(c.w, join)
		}
	})
}

func (c *formatter) add(s ...string) {
	for _, v := range s {
		fmt.Fprint(c.w, v)
	}
}

func (c *formatter) cleanExprs(list []ast.Expr, join string) {
	for i, n := range list {
		c.cleanCode(n)
		if i < len(list)-1 {
			c.add(join)
		}
	}
}

func (c *formatter) cleanStmts(list []ast.Stmt, join string) {
	for i, n := range list {
		c.cleanCode(n)
		if i < len(list)-1 {
			c.add(join)
		}
	}
}
func (c *formatter) cleanSpecs(list []ast.Spec, join string) {
	for i, n := range list {
		c.cleanCode(n)
		if i < len(list)-1 {
			c.add(join)
		}
	}
}

func (c *formatter) cleanIdents(list []*ast.Ident, join string) {
	for i, n := range list {
		c.add(n.Name)
		if i < len(list)-1 {
			c.add(join)
		}
	}
}

func (c *formatter) cleanFields(list []*ast.Field, join string) {
	for i, n := range list {
		c.cleanCode(n)
		if i < len(list)-1 {
			c.add(join)
		}
	}
}
