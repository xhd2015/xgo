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
		nameVal := name.Value
		line := name.Pos().Line()
		decls = append(decls, &DeclInfo{
			Kind: declKind,
			Name: nameVal,

			Line: int(line),
		})
	}
	return decls
}

func trapVariables(pkgPath string, fileList []*syntax.File, funcDelcs []*DeclInfo) {
	names := make(map[string]*DeclInfo, len(funcDelcs))
	varNames := make(map[string]*pkgdata.VarInfo)
	constNames := make(map[string]*pkgdata.ConstInfo)
	for _, funcDecl := range funcDelcs {
		identityName := funcDecl.IdentityName()
		names[identityName] = funcDecl
		if funcDecl.Kind == Kind_Var || funcDecl.Kind == Kind_VarPtr {
			varNames[identityName] = &pkgdata.VarInfo{
				Trap: funcDecl.FollowingTrapConst,
			}
		} else if funcDecl.Kind == Kind_Const {
			constDecl := funcDecl.ConstDecl
			constInfo := &pkgdata.ConstInfo{Untyped: true}
			if constDecl.Type != nil {
				constInfo.Untyped = false
			} else {
				constInfo.Type = getConstDeclValueType(constDecl.Values)
			}
			constNames[identityName] = constInfo
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
			// each function is a
			ctx := &BlockContext{
				Names: make(map[string]bool),
			}
			argNames := getFuncDeclNamesNoBlank(fnDecl.Recv, fnDecl.Type)
			for _, argName := range argNames {
				ctx.Names[argName] = true
			}
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

// a BlockContext provides a point where stmts can be prepended or inserted
type BlockContext struct {
	Parent *BlockContext
	Block  *syntax.BlockStmt

	Names map[string]bool

	// node appears as RHS of var decl
	ListExprParent       map[syntax.Node]*syntax.ListExpr
	RHSVarDeclParent     map[syntax.Node]*syntax.VarDecl
	OperationParent      map[syntax.Node]*syntax.Operation
	ArgCallExprParent    map[syntax.Node]*syntax.CallExpr
	RHSAssignNoDefParent map[syntax.Node]*syntax.AssignStmt
	CaseClauseParent     map[syntax.Node]*syntax.CaseClause
	ReturnStmtParent     map[syntax.Node]*syntax.ReturnStmt
	ParenParent          map[syntax.Node]*syntax.ParenExpr
	SelectorParent       map[syntax.Node]*syntax.SelectorExpr

	// const info
	ConstInfo map[syntax.Node]*ConstInfo

	// to be inserted
	ChildrenInsertList [][]syntax.Stmt

	TrapNames []*NameAndDecl
}

func (c *BlockContext) PrependStmtBeforeLastChild(stmt []syntax.Stmt) {
	n := len(c.ChildrenInsertList)
	c.ChildrenInsertList[n-1] = append(c.ChildrenInsertList[n-1], stmt...)
}

type ConstInfo struct {
	Type string
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

// avoid unused warning
var _ = (*BlockContext).traverseNode

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
		errorUnknown("node", node)

	}
	return node
}
func errorUnknown(expectType string, node syntax.Node) {
	// unknown
	if os.Getenv("XGO_DEBUG_VAR_TRAP_STRICT") == "true" {
		panic(fmt.Errorf("unrecognized %s: %T", expectType, node))
	}
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
		node.Call = ctx.traverseCallStmtCallExpr(node.Call, globaleNames, imports)
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
		if node.Results != nil {
			if ctx.ReturnStmtParent == nil {
				ctx.ReturnStmtParent = make(map[syntax.Node]*syntax.ReturnStmt, 1)
			}
			ctx.ReturnStmtParent[node.Results] = node
		}
		node.Results = ctx.traverseExpr(node.Results, globaleNames, imports)
	default:
		errorUnknown("stmt", node)
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
		} else {
			if ctx.RHSAssignNoDefParent == nil {
				ctx.RHSAssignNoDefParent = make(map[syntax.Node]*syntax.AssignStmt, 1)
			}
			ctx.RHSAssignNoDefParent[node.Rhs] = node
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
		errorUnknown("simple stmt", node)
	}
	return node
}

func (ctx *BlockContext) traverseBlockStmt(node *syntax.BlockStmt, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.BlockStmt {
	if node == nil {
		return nil
	}
	subCtx := &BlockContext{
		Parent: ctx,
	}
	n := len(node.List)
	for i := 0; i < n; i++ {
		subCtx.ChildrenInsertList = append(subCtx.ChildrenInsertList, nil)
		node.List[i] = subCtx.traverseStmt(node.List[i], globaleNames, imports)
	}
	for i := n - 1; i >= 0; i-- {
		node.List = insertBefore(node.List, i, subCtx.ChildrenInsertList[i])
	}
	return node
}

func (ctx *BlockContext) traverseCaseClause(node *syntax.CaseClause, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.CaseClause {
	if node == nil {
		return nil
	}
	if node.Cases != nil {
		if ctx.CaseClauseParent == nil {
			ctx.CaseClauseParent = make(map[syntax.Node]*syntax.CaseClause, 1)
		}
		ctx.CaseClauseParent[node.Cases] = node
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

func (c *BlockContext) recordParen(paren *syntax.ParenExpr) {
	if paren == nil || paren.X == nil {
		return
	}
	if c.ParenParent == nil {
		c.ParenParent = make(map[syntax.Node]*syntax.ParenExpr, 1)
	}
	c.ParenParent[paren.X] = paren
}
func (c *BlockContext) recordSelectorExpr(sel *syntax.SelectorExpr) {
	if sel == nil || sel.X == nil {
		return
	}
	if c.SelectorParent == nil {
		c.SelectorParent = make(map[syntax.Node]*syntax.SelectorExpr, 1)
	}
	c.SelectorParent[sel.X] = sel
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
		// add names of function declares
		funcCtx := &BlockContext{
			Parent: ctx,
			Names:  make(map[string]bool),
		}
		argNames := getFuncDeclNamesNoBlank(nil, node.Type)
		for _, argName := range argNames {
			funcCtx.Names[argName] = true
		}
		node.Body = funcCtx.traverseBlockStmt(node.Body, globaleNames, imports)
		return node
	case *syntax.ParenExpr:
		ctx.recordParen(node)
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		if xgoConv, ok := node.X.(*syntax.XgoSimpleConvert); ok {
			constType := getConstType(xgoConv)
			newNode := &syntax.XgoSimpleConvert{
				X: &syntax.CallExpr{
					Fun:     syntax.NewName(node.Pos(), constType),
					ArgList: []syntax.Expr{node},
				},
			}
			ctx.recordConstType(newNode, constType)
			return newNode
		}
	case *syntax.SelectorExpr:
		ctx.recordSelectorExpr(node)
		newNode, xIsName := ctx.trapSelector(node, node, false, globaleNames, imports)
		if newNode != nil {
			return newNode
		}
		if !xIsName {
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
		node.X = ctx.traverseExpr(node.X, globaleNames, imports)
		if node.Lhs != nil {
			ctx.Add(node.Lhs.Value)
		}
	case *syntax.Operation:
		// take addr?
		if node.Op == syntax.And && node.Y == nil {
			// &a,
			switch x := node.X.(type) {
			case *syntax.Name:
				return ctx.trapAddrNode(node, x, globaleNames)
			case *syntax.SelectorExpr:
				newNode, xIsName := ctx.trapSelector(node, x, true, globaleNames, imports)
				if newNode != nil {
					return newNode
				}
				if xIsName {
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
		// if both side are const, then the operation should also
		// be wrapped in a const if the operation is not ==
		if node.X != nil && node.Y != nil && !isBoolOp(node.Op) {
			xConst := ctx.ConstInfo[node.X]
			yConst := ctx.ConstInfo[node.Y]
			isXgoConv := func(node syntax.Expr) bool {
				_, ok := node.(*syntax.XgoSimpleConvert)
				return ok
			}
			// if both are constant,skip wrapping
			if xConst != nil && yConst != nil && (isXgoConv(node.X) || isXgoConv(node.Y)) {
				newNode := newConv(node, xConst.Type)
				ctx.recordConstType(newNode, xConst.Type)
				return newNode
			}
		} else if node.Y == nil && (node.Op == syntax.Add || node.Op == syntax.Sub) {
			if xgoConv, ok := node.X.(*syntax.XgoSimpleConvert); ok {
				constType := getConstType(xgoConv)
				newNode := newConv(node, constType)
				ctx.recordConstType(newNode, constType)
				return newNode
			}
		}
		return node
	case *syntax.CallExpr:
		return ctx.traverseCallExpr(node, globaleNames, imports)
	case *syntax.ListExpr:
		if len(node.ElemList) > 0 {
			if ctx.ListExprParent == nil {
				ctx.ListExprParent = make(map[syntax.Node]*syntax.ListExpr, len(node.ElemList))
			}
			for _, elem := range node.ElemList {
				ctx.ListExprParent[elem] = node
			}
		}
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
		constType := getBasicLitConstType(node.Kind)
		if constType != "" {
			if ctx.ConstInfo == nil {
				ctx.ConstInfo = make(map[syntax.Node]*ConstInfo, 1)
			}
			ctx.ConstInfo[node] = &ConstInfo{Type: constType}
		}
	case *syntax.BadExpr:
	default:
		errorUnknown("expr", node)
	}
	return node
}
func isBoolOp(op syntax.Operator) bool {
	switch op {
	case syntax.Eql, syntax.Neq, syntax.Lss, syntax.Leq, syntax.Gtr, syntax.Geq:
		return true
	}
	return false
}

func getConstType(xgoConv *syntax.XgoSimpleConvert) string {
	return xgoConv.X.(*syntax.CallExpr).Fun.(*syntax.Name).Value
}

func newConv(node syntax.Expr, constType string) *syntax.XgoSimpleConvert {
	return &syntax.XgoSimpleConvert{
		X: &syntax.CallExpr{
			Fun:     syntax.NewName(node.Pos(), constType),
			ArgList: []syntax.Expr{node},
		},
	}
}

func (ctx *BlockContext) recordConstType(node syntax.Node, constType string) {
	if ctx.ConstInfo == nil {
		ctx.ConstInfo = make(map[syntax.Node]*ConstInfo, 1)
	}
	ctx.ConstInfo[node] = &ConstInfo{Type: constType}
}

func (ctx *BlockContext) traverseCallExpr(node *syntax.CallExpr, globaleNames map[string]*DeclInfo, imports map[string]string) *syntax.CallExpr {
	if node == nil {
		return nil
	}
	if ctx.ArgCallExprParent == nil {
		ctx.ArgCallExprParent = make(map[syntax.Node]*syntax.CallExpr, len(node.ArgList))
		for _, arg := range node.ArgList {
			ctx.ArgCallExprParent[arg] = node
		}
	}

	// NOTE: previousely we skip capturing a name as a function(i.e. the next statement is commented out)
	// reason: we cannot tell if the receiver is
	// a pointer or a value. If the target function requires a
	// pointer while we assign a value, then effect will lost, such
	// as lock and unlock.
	//
	// however, since we have isVarOKToTrap() to check if a variable is ok to trap, so this can be turned on, resulting in workaround:
	//    A.B() -> typeOfA(A).B() which makes things work

	node.Fun = ctx.traverseExpr(node.Fun, globaleNames, imports)
	for i, arg := range node.ArgList {
		node.ArgList[i] = ctx.traverseExpr(arg, globaleNames, imports)
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
		// var a int64 = N
		if node.Values != nil {
			if ctx.RHSVarDeclParent == nil {
				ctx.RHSVarDeclParent = make(map[syntax.Node]*syntax.VarDecl, 1)
			}
			ctx.RHSVarDeclParent[node.Values] = node
			node.Values = ctx.traverseExpr(node.Values, globaleNames, imports)
		}
	default:
		errorUnknown("stmt", node)
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
	var explicitType syntax.Expr
	var rhsAssign *syntax.AssignStmt
	var isCallArg bool
	var untypedConstType string
	if decl.Kind == Kind_Var || decl.Kind == Kind_VarPtr {
		if !decl.FollowingTrapConst && !c.isVarOKToTrap(node) {
			return node
		}
		// good to go
	} else if decl.Kind == Kind_Const {
		// untyped const(most cases) should only be used in
		// several cases because runtime type is unknown
		if decl.ConstDecl.Type == nil {
			if !xgo_ctxt.EnableTrapUntypedConst {
				return node
			}
			untypedConstType = getConstDeclValueType(decl.ConstDecl.Values)
			if untypedConstType == "" {
				return node
			}
			var ok bool
			explicitType, ok = c.isConstOKToTrap(node)
			if !ok {
				// debug
				if _, ok := c.ArgCallExprParent[node]; ok {
					isCallArg = true
				}
				if !isCallArg {
					return node
				}
			}
		}
	} else {
		return node
	}
	preStmts, varDefStmt, tmpVarName := trapVar(node, syntax.NewName(node.Pos(), XgoLocalPkgName), node.Value, false)
	if rhsAssign != nil {
		varDefStmt.Op = 0
		preStmts = append([]syntax.Stmt{
			&syntax.AssignStmt{
				Op:  syntax.Def,
				Lhs: syntax.NewName(node.Pos(), tmpVarName),
				Rhs: rhsAssign.Lhs,
			},
		}, preStmts...)
	}

	c.PrependStmtBeforeLastChild(preStmts)
	newName := syntax.NewName(node.Pos(), tmpVarName)
	if explicitType != nil {
		return &syntax.CallExpr{
			Fun: explicitType,
			ArgList: []syntax.Expr{
				newName,
			},
		}
	}
	if untypedConstType != "" {
		newNode := newConv(newName, untypedConstType)
		c.recordConstType(newNode, untypedConstType)
		c.ConstInfo[newNode] = &ConstInfo{Type: untypedConstType}
		return newNode
	}
	return newName
}

func (ctx *BlockContext) trapSelector(node syntax.Expr, sel *syntax.SelectorExpr, takeAddr bool, globaleNames map[string]*DeclInfo, imports map[string]string) (newExpr syntax.Expr, xIsName bool) {
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
	var explicitType syntax.Expr
	pkgData := pkgdata.GetPkgData(pkgPath)
	var isCallArg bool
	var untypedConstType string
	if constInfo, ok := pkgData.Consts[sel.Sel.Value]; ok {
		if constInfo.Untyped {
			if !xgo_ctxt.EnableTrapUntypedConst {
				return nil, true
			}
			untypedConstType = constInfo.Type
			if untypedConstType == "" {
				return nil, true
			}
			var ok bool
			explicitType, ok = ctx.isConstOKToTrap(node)
			if !ok {
				// debug
				if _, ok := ctx.ArgCallExprParent[node]; ok {
					isCallArg = true
				}
				if !isCallArg {
					return nil, true
				}

			}
		}
	} else if varInfo, ok := pkgData.Vars[sel.Sel.Value]; ok {
		if !varInfo.Trap && !takeAddr && !ctx.isVarOKToTrap(node) {
			return nil, true
		}
	} else {
		return nil, true
	}
	preStmts, _, tmpVarName := trapVar(node, newStringLit(pkgPath), sel.Sel.Value, takeAddr)
	ctx.PrependStmtBeforeLastChild(preStmts)
	newName := syntax.NewName(node.Pos(), tmpVarName)
	if explicitType != nil {
		return &syntax.CallExpr{
			Fun: explicitType,
			ArgList: []syntax.Expr{
				newName,
			},
		}, true
	}
	if untypedConstType != "" {
		newNode := newConv(newName, untypedConstType)
		ctx.recordConstType(newNode, untypedConstType)
		return newNode, true
	}
	return newName, true
}

// check if node is either X of an X.Y, or (X).Y or ((X)).Y...
func (ctx *BlockContext) isVarOKToTrap(node syntax.Node) bool {
	// a variable can only trapped when it will not
	// cause an implicit pointer
	_, ok := ctx.SelectorParent[node]
	if ok {
		return false
	}
	paren, ok := ctx.ParenParent[node]
	if ok {
		return ctx.isVarOKToTrap(paren)
	}
	return true
}

func (ctx *BlockContext) isConstOKToTrap(node syntax.Node) (explicitType syntax.Expr, ok bool) {
	if true {
		return nil, true
	}
	// is const and inside operation
	if _, ok := ctx.OperationParent[node]; ok {
		return nil, false
	}

	// NOTE: will this: int64(a) not work? maybe we
	// can make it work
	if _, ok := ctx.ArgCallExprParent[node]; ok {
		// directly as argument to a call
		return nil, false
	}
	if _, ok := ctx.CaseClauseParent[node]; ok {
		return nil, false
	}
	if _, ok := ctx.ReturnStmtParent[node]; ok {
		return nil, false
	}
	if varDecl, ok := ctx.RHSVarDeclParent[node]; ok {
		return varDecl.Type, true
	}

	if _, ok := ctx.RHSAssignNoDefParent[node]; ok {
		// a=CONST -> tmp:=CONST,a=tmp
		// not working
		// rhsAssign = assign
		return nil, false
	}
	listExprParent, ok := ctx.ListExprParent[node]
	if !ok {
		return nil, true
	}
	return ctx.isConstOKToTrap(listExprParent)
}

func getConstDeclValueType(expr syntax.Expr) string {
	switch expr := expr.(type) {
	case *syntax.BasicLit:
		return getBasicLitConstType(expr.Kind)
	case *syntax.Name:
		if expr.Value == "true" || expr.Value == "false" {
			return "bool"
		}
		// NOTE: nil is not a constant
	case *syntax.Operation:
		// binary operation
		if isBoolOp(expr.Op) {
			return "bool"
		}
		xIsNil := expr.X == nil
		yIsNil := expr.Y == nil
		if xIsNil && yIsNil {
			return ""
		}
		// see https://github.com/xhd2015/xgo/issues/172
		// var x = 5*y, y's type is not known
		var xType string
		if !xIsNil {
			xType = getConstDeclValueType(expr.X)
		}
		var yType string
		if !yIsNil {
			yType = getConstDeclValueType(expr.Y)
		}
		// either is nil
		if xIsNil != yIsNil {
			if xType != "" {
				return xType
			}
			return yType
		}
		// 5*5 -> int
		// 5*X -> unknown
		if xType == yType {
			return xType
		}
		return ""
		// TODO: handle SelectorExpr and iota
	case *syntax.ParenExpr:
		return getConstDeclValueType(expr.X)
	}
	return ""
}
func getBasicLitConstType(kind syntax.LitKind) string {
	switch kind {
	case syntax.IntLit:
		return "int"
	case syntax.StringLit:
		return "string"
	case syntax.RuneLit:
		return "rune"
	case syntax.FloatLit:
		return "float64"
	}
	return ""
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
	preStmts, _, tmpVarName := trapVar(node, syntax.NewName(nameNode.Pos(), XgoLocalPkgName), name, true)
	c.PrependStmtBeforeLastChild(preStmts)
	return syntax.NewName(node.Pos(), tmpVarName)
}

func trapVar(expr syntax.Expr, pkgRef syntax.Expr, name string, takeAddr bool) (preStmts []syntax.Stmt, varDefStmt *syntax.AssignStmt, tmpVarName string) {
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
	varDefStmt = &syntax.AssignStmt{
		Op:  syntax.Def,
		Lhs: syntax.NewName(pos, varName),
		Rhs: expr,
	}

	preStmts = append(preStmts,
		varDefStmt,
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
	return preStmts, varDefStmt, varName

}

func insertBefore(list []syntax.Stmt, i int, add []syntax.Stmt) []syntax.Stmt {
	if len(add) == 0 {
		return list
	}
	return append(append(list[:i:i], add...), list[i:]...)
}
