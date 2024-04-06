package syntax

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	xgo_ctxt "cmd/compile/internal/xgo_rewrite_internal/patch/ctxt"
	"cmd/compile/internal/xgo_rewrite_internal/patch/pkgdata"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func allowVarTrap() bool {
	pkgPath := xgo_ctxt.GetPkgPath()
	return allowPkgVarTrap(pkgPath)
}

func allowPkgVarTrap(pkgPath string) bool {
	// prevent all std variables
	if base.Flag.Std {
		return false
	}
	mainModule := xgo_ctxt.XgoMainModule
	if mainModule == "" {
		return false
	}

	if strings.HasPrefix(pkgPath, mainModule) && (len(pkgPath) == len(mainModule) || pkgPath[len(mainModule)] == '/') {
		return true
	}
	return false
}

func collectVarDecls(declKind DeclKind, names []*syntax.Name, typ syntax.Expr) []*DeclInfo {
	var decls []*DeclInfo
	for _, name := range names {
		line := name.Pos().Line()
		decls = append(decls, &DeclInfo{
			Kind: declKind,
			Name: name.Value,

			Line: int(line),
		})
	}
	return decls
}

type vis struct {
}

var _ syntax.Visitor = (*vis)(nil)

// Visit implements syntax.Visitor.
func (c *vis) Visit(node syntax.Node) (w syntax.Visitor) {
	return nil
}

func trapVariables(pkgPath string, fileList []*syntax.File, funcDelcs []*DeclInfo) {
	names := make(map[string]*DeclInfo, len(funcDelcs))
	varNames := make(map[string]bool)
	constNames := make(map[string]bool)
	for _, funcDecl := range funcDelcs {
		identityName := funcDecl.IdentityName()
		names[identityName] = funcDecl
		if funcDecl.Kind == Kind_Var || funcDecl.Kind == Kind_VarPtr {
			varNames[identityName] = true
		} else if funcDecl.Kind == Kind_Const {
			constNames[identityName] = true
		}
	}
	err := pkgdata.WritePkgData(pkgPath, &pkgdata.PackageData{
		Consts: constNames,
		Vars:   varNames,
	})
	if err != nil {
		base.Fatalf("write pkg data: %v", err)
	}
	// iterate each file, find variable reference,
	for _, file := range fileList {
		imports := getImports(file)
		for _, decl := range file.DeclList {
			fnDecl, ok := decl.(*syntax.FuncDecl)
			if !ok {
				continue
			}
			if fnDecl.Body == nil {
				continue
			}
			ctx := &BlockContext{}
			fnDecl.Body = ctx.traverseBlockStmt(fnDecl.Body, names, imports)
		}
	}
}
func getImports(file *syntax.File) map[string]string {
	imports := make(map[string]string)
	for _, decl := range file.DeclList {
		impDecl, ok := decl.(*syntax.ImportDecl)
		if !ok {
			continue
		}
		pkgPath, err := strconv.Unquote(impDecl.Path.Value)
		if err != nil {
			continue
		}
		var localName string
		if impDecl.LocalPkgName != nil {
			localName = impDecl.LocalPkgName.Value
		} else {
			idx := strings.LastIndex(pkgPath, "/")
			if idx < 0 {
				localName = pkgPath
			} else {
				localName = pkgPath[idx+1:]
			}
		}
		if localName == "" || localName == "." || localName == "_" {
			continue
		}
		imports[localName] = pkgPath
	}
	return imports
}

type BlockContext struct {
	Parent *BlockContext
	Block  *syntax.BlockStmt
	Index  int

	Children []*BlockContext

	Names map[string]bool

	OperationParent map[syntax.Node]*syntax.Operation

	// to be inserted
	InsertList []syntax.Stmt

	TrapNames []*NameAndDecl
}

type NameAndDecl struct {
	TakeAddr bool
	Name     *syntax.Name
	Decl     *DeclInfo
}

func (c *BlockContext) Add(name string) {
	if c.Names == nil {
		c.Names = make(map[string]bool, 1)
	}
	c.Names[name] = true
}
func (c *BlockContext) Has(name string) bool {
	if c == nil {
		return false
	}
	_, ok := c.Names[name]
	if ok {
		return true
	}
	return c.Parent.Has(name)
}

// imports: name -> pkgPath
func (ctx *BlockContext) traverseNode(node syntax.Node, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.Node {
	if node == nil {
		return nil
	}
	switch node := node.(type) {
	case syntax.Stmt:
		return ctx.traverseStmt(node, globaleNames, imports)
	case syntax.Expr:
		return ctx.traverseExpr(node, globaleNames, imports)
	case *syntax.CaseClause:
		return ctx.traverseCaseClause(node, globaleNames, imports)
	case *syntax.CommClause:
		return ctx.traverseCommonClause(node, globaleNames, imports)
	case *syntax.Field:
		// ignore
	default:
		// unknown
		if os.Getenv("XGO_DEBUG_VAR_TRAP_LOOSE") != "true" {
			panic(fmt.Errorf("unrecognized node: %T", node))
		}
	}
	return node
}

func (ctx *BlockContext) traverseStmt(node syntax.Stmt, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.Stmt {
	if node == nil {
		return nil
	}
	switch node := node.(type) {
	case syntax.SimpleStmt:
		return ctx.traverseSimpleStmt(node, globaleNames, imports)
	case *syntax.BlockStmt:
		return ctx.traverseBlockStmt(node, globaleNames, imports)
	case *syntax.CallStmt:
		// defer, go
		node.Call = ctx.traverseExpr(node.Call, globaleNames, imports)
		return node
	case *syntax.IfStmt:
		node.Init = ctx.traverseSimpleStmt(node.Init, globaleNames, imports)
		node.Cond = ctx.traverseExpr(node.Cond, globaleNames, imports)
		node.Then = ctx.traverseBlockStmt(node.Then, globaleNames, imports)
		node.Else = ctx.traverseStmt(node.Else, globaleNames, imports)
		return node
	case *syntax.ForStmt:
		node.Init = ctx.traverseSimpleStmt(node.Init, globaleNames, imports)
		node.Cond = ctx.traverseExpr(node.Cond, globaleNames, imports)
		node.Post = ctx.traverseSimpleStmt(node.Post, globaleNames, imports)
		node.Body = ctx.traverseBlockStmt(node.Body, globaleNames, imports)
	case *syntax.SwitchStmt:
		node.Init = ctx.traverseSimpleStmt(node.Init, globaleNames, imports)
		node.Tag = ctx.traverseExpr(node.Tag, globaleNames, imports)
		for i, clause := range node.Body {
			node.Body[i] = ctx.traverseCaseClause(clause, globaleNames, imports)
		}
	case *syntax.SelectStmt:
		for i, clause := range node.Body {
			node.Body[i] = ctx.traverseCommonClause(clause, globaleNames, imports)
		}
	case *syntax.DeclStmt:
		for i, decl := range node.DeclList {
			node.DeclList[i] = ctx.traverseDecl(decl, globaleNames, imports)
		}
	case *syntax.LabeledStmt:
		node.Stmt = ctx.traverseStmt(node.Stmt, globaleNames, imports)
	case *syntax.BranchStmt:
		// ignore continue or continue label
	case *syntax.ReturnStmt:
		node.Results = ctx.traverseExpr(node.Results, globaleNames, imports)
	default:
		// unknown
		if os.Getenv("XGO_DEBUG_VAR_TRAP_LOOSE") != "true" {
			panic(fmt.Errorf("unrecognized stmt: %T", node))
		}
	}
	return node
}

func (ctx *BlockContext) traverseSimpleStmt(node syntax.SimpleStmt, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.SimpleStmt {
	if node == nil {
		return nil
	}
	switch node := node.(type) {
	case *syntax.ExprStmt:
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		return node
	case *syntax.SendStmt:
		node.Chan = ctx.traverseExpr(node.Chan, globaleNames, imports)
		node.Value = ctx.traverseExpr(node.Value, globaleNames, imports)
	case *syntax.AssignStmt:
		if node.Op == syntax.Def {
			// add name to current scope
			if name, ok := node.Lhs.(*syntax.Name); ok {
				ctx.Add(name.Value)
			} else if names, ok := node.Lhs.(*syntax.ListExpr); ok {
				for _, elem := range names.ElemList {
					if name, ok := elem.(*syntax.Name); ok {
						ctx.Add(name.Value)
					}
				}
			}
		}
		node.Rhs = ctx.traverseExpr(node.Rhs, globaleNames, imports)
	case *syntax.RangeClause:
		if node.Lhs != nil && node.Def {
			var fakeAssign syntax.Stmt = &syntax.AssignStmt{
				Op:  syntax.Def,
				Lhs: node.Lhs,
			}
			ctx.traverseStmt(fakeAssign, globaleNames, imports)
		}
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
	case *syntax.EmptyStmt:
		// nothing
	default:
		// unknown
		if os.Getenv("XGO_DEBUG_VAR_TRAP_LOOSE") != "true" {
			panic(fmt.Errorf("unrecognized simple stmt: %T", node))
		}
	}
	return node
}

func (ctx *BlockContext) traverseBlockStmt(node *syntax.BlockStmt, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.BlockStmt {
	if node == nil {
		return nil
	}
	n := len(node.List)
	for i := 0; i < n; i++ {
		subCtx := &BlockContext{
			Parent: ctx,
			Block:  node,
			Index:  i,
		}
		ctx.Children = append(ctx.Children, subCtx)
		node.List[i] = subCtx.traverseStmt(node.List[i], globaleNames, imports)
	}
	for i := n - 1; i >= 0; i-- {
		node.List = insertBefore(node.List, i, ctx.Children[i].InsertList)
	}
	return node
}

func (ctx *BlockContext) traverseCaseClause(node *syntax.CaseClause, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.CaseClause {
	if node == nil {
		return nil
	}
	node.Cases = ctx.traverseExpr(node.Cases, globaleNames, imports)
	fakeBlock := &syntax.BlockStmt{
		List: node.Body,
	}
	fakeBlock = ctx.traverseBlockStmt(fakeBlock, globaleNames, imports)
	node.Body = fakeBlock.List
	return node
}

func (ctx *BlockContext) traverseCommonClause(node *syntax.CommClause, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.CommClause {
	if node == nil {
		return nil
	}
	node.Comm = ctx.traverseSimpleStmt(node.Comm, globaleNames, imports)
	fakeBlock := &syntax.BlockStmt{
		List: node.Body,
	}
	fakeBlock = ctx.traverseBlockStmt(fakeBlock, globaleNames, imports)
	node.Body = fakeBlock.List
	return node
}

func (ctx *BlockContext) traverseExpr(node syntax.Expr, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.Expr {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *syntax.Name:
		return ctx.trapValueNode(node, globaleNames)
	case *syntax.CompositeLit:
		for i, e := range node.ElemList {
			node.ElemList[i] = ctx.traverseExpr(e, globaleNames, imports)
		}
	case *syntax.KeyValueExpr:
		node.Value = ctx.traverseExpr(node.Value, globaleNames, imports)
	case *syntax.FuncLit:
		subCtx := &BlockContext{
			Parent: ctx,
		}
		// TODO: add names of types
		ctx.Children = append(ctx.Children, subCtx)
		node.Body = subCtx.traverseBlockStmt(node.Body, globaleNames, imports)
		return node
	case *syntax.ParenExpr:
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
	case *syntax.SelectorExpr:
		newNode, selIsName := ctx.trapSelector(node, node, false, globaleNames, imports)
		if newNode != nil {
			return newNode
		}
		if !selIsName {
			node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		}
	case *syntax.IndexExpr:
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		node.Index = ctx.traverseExpr(node.Index, globaleNames, imports)
	case *syntax.SliceExpr:
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		for i := 0; i < len(node.Index); i++ {
			node.Index[i] = ctx.traverseExpr(node.Index[i], globaleNames, imports)
		}
	case *syntax.AssertExpr:
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
	case *syntax.TypeSwitchGuard:
		res := ctx.traverseExpr(node.X, globaleNames, imports)
		if node.Lhs != nil {
			ctx.Add(node.Lhs.Value)
		}
		return res
	case *syntax.Operation:
		// take addr?
		if node.Op == syntax.And && node.Y == nil {
			// &a,
			switch x := node.X.(type) {
			case *syntax.Name:
				return ctx.trapAddrNode(node, x, globaleNames)
			case *syntax.SelectorExpr:
				newNode, selIsName := ctx.trapSelector(node, x, true, globaleNames, imports)
				if newNode != nil {
					return newNode
				}
				if selIsName {
					return node
				}
			}
		}
		if node.X != nil && node.Y != nil {
			if ctx.OperationParent == nil {
				ctx.OperationParent = make(map[syntax.Node]*syntax.Operation)
			}
			ctx.OperationParent[node.X] = node
			ctx.OperationParent[node.Y] = node
		}
		// x op y
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		node.Y = ctx.traverseExpr(node.Y, globaleNames, imports)
		return node
	case *syntax.CallExpr:
		// NOTE: we skip capturing a name as a function
		// node.Fun = ctx.traverseExpr(node.Fun, globaleNames, imports)
		for i, arg := range node.ArgList {
			node.ArgList[i] = ctx.traverseExpr(arg, globaleNames, imports)
		}
	case *syntax.ListExpr:
		for i, elem := range node.ElemList {
			node.ElemList[i] = ctx.traverseExpr(elem, globaleNames, imports)
		}
		// the following are ignored
	case *syntax.ArrayType:
	case *syntax.SliceType:
	case *syntax.DotsType:
	case *syntax.StructType:
	case *syntax.InterfaceType:
	case *syntax.FuncType:
	case *syntax.ChanType:
	case *syntax.MapType:
	case *syntax.BasicLit:
	case *syntax.BadExpr:
	default:
		// unknown
		if os.Getenv("XGO_DEBUG_VAR_TRAP_LOOSE") != "true" {
			panic(fmt.Errorf("unrecognized expr: %T", node))
		}
	}
	return node
}

func (ctx *BlockContext) traverseDecl(node syntax.Decl, globaleNames map[string]*DeclInfo, imports map[string]string) syntax.Decl {
	if node == nil {
		return nil
	}
	switch node := node.(type) {
	case *syntax.ConstDecl:
	case *syntax.TypeDecl:
	case *syntax.VarDecl:
	default:
		// unknown
		if os.Getenv("XGO_DEBUG_VAR_TRAP_LOOSE") != "true" {
			panic(fmt.Errorf("unrecognized stmt: %T", node))
		}
	}
	return node
}

func (c *BlockContext) trapValueNode(node *syntax.Name, globaleNames map[string]*DeclInfo) syntax.Expr {
	name := node.Value
	if c.Has(name) {
		return node
	}
	// TODO: what about dot import?
	decl := globaleNames[name]
	if decl == nil {
		return node
	}
	if decl.Kind == Kind_Var || decl.Kind == Kind_VarPtr {
		// good to go
	} else if decl.Kind == Kind_Const {
		if _, ok := c.OperationParent[node]; ok {
			// directly inside an operation
			return node
		}
	} else {
		return node
	}
	preStmts, tmpVarName := trapVar(node, syntax.NewName(node.Pos(), XgoLocalPkgName), node.Value, false)
	c.InsertList = append(c.InsertList, preStmts...)
	return syntax.NewName(node.Pos(), tmpVarName)
}

func (ctx *BlockContext) trapSelector(node syntax.Expr, sel *syntax.SelectorExpr, takeAddr bool, globaleNames map[string]*DeclInfo, imports map[string]string) (newExpr syntax.Expr, selIsName bool) {
	// form: pkg.var
	nameNode, ok := sel.X.(*syntax.Name)
	if !ok {
		return nil, false
	}
	name := nameNode.Value
	if ctx.Has(name) {
		// local name
		return nil, true
	}
	// import path
	pkgPath := imports[name]
	if pkgPath == "" {
		sel.X = ctx.trapValueNode(nameNode, globaleNames)
		return nil, true
	}
	if !allowPkgVarTrap(pkgPath) {
		return nil, true
	}
	pkgData := pkgdata.GetPkgData(pkgPath)
	if pkgData.Consts[sel.Sel.Value] {
		// is const and inside operation
		if _, ok := ctx.OperationParent[node]; ok {
			return nil, true
		}
	}
	preStmts, tmpVarName := trapVar(node, newStringLit(pkgPath), sel.Sel.Value, takeAddr)
	ctx.InsertList = append(ctx.InsertList, preStmts...)
	return syntax.NewName(node.Pos(), tmpVarName), true
}

func (c *BlockContext) trapAddrNode(node *syntax.Operation, nameNode *syntax.Name, globaleNames map[string]*DeclInfo) syntax.Expr {
	name := nameNode.Value
	if c.Has(name) {
		return node
	}
	// TODO: what about dot import?
	decl := globaleNames[name]
	if decl == nil || !decl.Kind.IsVarOrConst() {
		return node
	}
	preStmts, tmpVarName := trapVar(node, syntax.NewName(nameNode.Pos(), XgoLocalPkgName), name, true)
	c.InsertList = append(c.InsertList, preStmts...)
	return syntax.NewName(node.Pos(), tmpVarName)
}

func trapVar(expr syntax.Expr, pkgRef syntax.Expr, name string, takeAddr bool) (preStmts []syntax.Stmt, tmpVarName string) {
	pos := expr.Pos()
	line := pos.Line()
	col := pos.Col()

	// a.b:
	// ___m := a;__trap_var(&__m, &__a);
	// __m.b

	// &a:
	//  __m:=&a; __trap_var(pkg,"a", &__m,takeAddr=true)
	//  &a -> __m
	varName := fmt.Sprintf("__xgo_%s_%d_%d", name, line, col)
	// a:

	preStmts = append(preStmts, &syntax.AssignStmt{
		Op:  syntax.Def,
		Lhs: syntax.NewName(pos, varName),
		Rhs: expr,
	},
		&syntax.ExprStmt{
			X: &syntax.CallExpr{
				Fun: syntax.NewName(pos, "__xgo_link_trap_var_for_generated"),
				ArgList: []syntax.Expr{
					pkgRef,
					newStringLit(name),
					&syntax.Operation{
						Op: syntax.And,
						X:  syntax.NewName(pos, varName),
					},
					newBool(pos, takeAddr),
				},
			},
		},
		// &syntax.ExprStmt{
		// 	X: &syntax.CallExpr{
		// 		Fun: syntax.NewName(pos, "panic"),
		// 		ArgList: []syntax.Expr{
		// 			newStringLit(fmt.Sprintf("%s := %s; __trap_var(&%s, &%s)", varName, name.Value, varName, name.Value)),
		// 		},
		// 	},
		// },
	)

	for _, preStmt := range preStmts {
		fillPos(pos, preStmt)
	}
	return preStmts, varName

}

func insertBefore(list []syntax.Stmt, i int, add []syntax.Stmt) []syntax.Stmt {
	return append(append(list[:i:i], add...), list[i:]...)
}
