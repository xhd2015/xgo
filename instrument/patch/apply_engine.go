package patch

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// cursor represents a position within a Go source file.
type cursor struct {
	offset    int  // byte offset for insert_before
	endOffset int  // byte offset for insert_after / end of replace range
	isReplace bool // set by find_for_replace
}

// edit represents a single edit operation collected during patch application.
type edit struct {
	seq       int    // sequence number (incremented per positioning command)
	offset    int    // insertion offset
	oldEnd    int    // for replace: end offset of old content
	newText   string // the text to insert/replace with
	oldText   string // for replace: the original text (for marker old tag)
	isReplace bool
}

// editGroup represents all edits at a single sequence number (cursor position).
type editGroup struct {
	seq       int
	offset    int    // insertion point for insert_before
	insertEnd int    // insertion point for insert_after
	segments  []string
	isReplace bool
	oldText   string // for replace: the original text
}

// applyPatch applies a single PatchBlock to the original Go source text.
func applyPatch(original string, block PatchBlock) (string, error) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "", original, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse target file: %w", err)
	}

	state := &applyState{
		original: original,
		fset:     fset,
		astFile:  astFile,
		groups:   make(map[int]*editGroup),
	}

	for _, cmd := range block.Commands {
		switch cmd.Type {
		case CmdGoto:
			c, err := evalGoto(state, cmd.GotoTarget)
			if err != nil {
				return "", fmt.Errorf("goto %s: %w", cmd.GotoTarget, err)
			}
			state.cursor = c
			state.seq++
			state.lastInsertMode = insertBefore

		case CmdMatch:
			c, err := evalMatch(state, cmd.SearchText, false)
			if err != nil {
				return "", fmt.Errorf("match %q: %w", cmd.SearchText, err)
			}
			state.cursor = c
			state.seq++
			state.lastInsertMode = insertBefore

		case CmdFindForReplace:
			c, err := evalMatch(state, cmd.SearchText, true)
			if err != nil {
				return "", fmt.Errorf("find_for_replace %q: %w", cmd.SearchText, err)
			}
			state.cursor = c
			state.seq++
			state.lastInsertMode = insertBefore

		case CmdInsertBefore:
			if cmd.EditText == "" {
				return "", fmt.Errorf("insert_before requires text")
			}
			state.lastInsertMode = insertBefore
			state.addSegment(cmd.EditText)

		case CmdInsertAfter:
			if cmd.EditText == "" {
				return "", fmt.Errorf("insert_after requires text")
			}
			state.lastInsertMode = insertAfter
			state.addSegment(cmd.EditText)

		case CmdReplace:
			if cmd.EditText == "" {
				return "", fmt.Errorf("replace requires text")
			}
			if !state.cursor.isReplace {
				return "", fmt.Errorf("replace requires prior find_for_replace")
			}
			oldText := state.original[state.cursor.offset:state.cursor.endOffset]
			state.lastInsertMode = insertBefore
			state.addReplace(cmd.EditText, oldText)

		case CmdNewline:
			state.addNewline()
		}
	}

	return state.applyEdits(block.Name), nil
}

type insertMode int

const (
	insertBefore insertMode = iota
	insertAfter
)

type applyState struct {
	original       string
	fset           *token.FileSet
	astFile        *ast.File
	cursor         cursor
	seq            int
	lastInsertMode insertMode

	groups map[int]*editGroup
}

func (s *applyState) getGroup() *editGroup {
	g, ok := s.groups[s.seq]
	if !ok {
		insertEnd := s.cursor.offset
		if s.cursor.endOffset > insertEnd {
			insertEnd = s.cursor.endOffset
		}
		g = &editGroup{
			seq:       s.seq,
			offset:    s.cursor.offset,
			insertEnd: insertEnd,
		}
		s.groups[s.seq] = g
	}
	return g
}

func (s *applyState) addSegment(text string) {
	g := s.getGroup()
	g.segments = append(g.segments, text)
}

func (s *applyState) addReplace(newText, oldText string) {
	g := s.getGroup()
	g.isReplace = true
	g.oldText = oldText
	g.segments = append(g.segments, newText)
}

func (s *applyState) addNewline() {
	g := s.getGroup()
	// newline merges with the previous segment: "A" + newline → "A\n"
	if len(g.segments) > 0 {
		g.segments[len(g.segments)-1] += "\n"
	}
}

// combineSegments combines text segments for an edit group.
// For insert_before: segments are collected in insertion order, then applied
// in reverse so the last written goes closest to original content.
// For insert_after: segments are applied in forward order.
// newline merges "\n" with the adjacent text segment, so it stays on the
// correct side of the text in the final output.
func (s *applyState) combineSegments(g *editGroup) string {
	if g.isReplace {
		return strings.Join(g.segments, "")
	}

	if len(g.segments) == 0 {
		return ""
	}

	var buf strings.Builder
	if s.lastInsertMode == insertBefore {
		for i := len(g.segments) - 1; i >= 0; i-- {
			buf.WriteString(g.segments[i])
		}
	} else {
		for _, seg := range g.segments {
			buf.WriteString(seg)
		}
	}

	return buf.String()
}

// applyEdits applies all collected edits and wraps them with markers.
func (s *applyState) applyEdits(blockName string) string {
	result := s.original

	// Process edits in REVERSE offset order (highest offset first)
	// to avoid position shifts when inserting text.
	maxSeq := s.seq
	for seq := maxSeq; seq >= 1; seq-- {
		g, ok := s.groups[seq]
		if !ok {
			continue
		}

		insertText := s.combineSegments(g)
		if insertText == "" && !g.isReplace {
			continue
		}

		if g.isReplace {
			oldTag := ""
			if g.oldText != "" {
				oldTag = "<old:" + g.oldText + ">"
			}
			markerBegin := fmt.Sprintf("/*<%s:%d%s>*/", blockName, seq, oldTag)
			markerEnd := "/*<end>*/"
			wrapped := markerBegin + insertText + markerEnd

			// Replace old content (from offset to oldEnd) with wrapped text
			result = result[:g.offset] + wrapped + result[g.offset+len(g.oldText):]
		} else {
			markerBegin := fmt.Sprintf("/*<%s:%d>*/", blockName, seq)
			markerEnd := "/*<end>*/"

			if s.lastInsertMode == insertBefore || seq < maxSeq {
				// insert_before: insert at offset
				wrapped := markerBegin + insertText + markerEnd
				result = result[:g.offset] + wrapped + result[g.offset:]
			} else {
				// insert_after: insert at insertEnd
				wrapped := markerBegin + insertText + markerEnd
				result = result[:g.insertEnd] + wrapped + result[g.insertEnd:]
			}
		}
	}

	return result
}

// evalGoto evaluates a goto command and returns a cursor positioned at the target.
func evalGoto(state *applyState, target string) (cursor, error) {
	switch {
	case target == "opening {":
		return evalGotoBrace(state, true)

	case target == "closing }":
		return evalGotoBrace(state, false)

	case strings.HasPrefix(target, "field "):
		fieldName := strings.TrimSpace(target[len("field "):])
		return evalGotoField(state, fieldName)

	case strings.HasPrefix(target, "struct "):
		name := strings.TrimSpace(target[len("struct "):])
		return evalGotoDecl(state, name, isStructType)

	case strings.HasPrefix(target, "interface "):
		name := strings.TrimSpace(target[len("interface "):])
		return evalGotoDecl(state, name, isInterfaceType)

	case strings.HasPrefix(target, "func "):
		rest := strings.TrimSpace(target[len("func "):])
		return evalGotoFunc(state, rest)

	default:
		return cursor{}, fmt.Errorf("unknown goto target: %q", target)
	}
}

func isStructType(ts *ast.TypeSpec) bool {
	_, ok := ts.Type.(*ast.StructType)
	return ok
}

func isInterfaceType(ts *ast.TypeSpec) bool {
	_, ok := ts.Type.(*ast.InterfaceType)
	return ok
}

func evalGotoDecl(state *applyState, name string, typeCheck func(*ast.TypeSpec) bool) (cursor, error) {
	for _, decl := range state.astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if ts.Name.Name == name && typeCheck(ts) {
				pos := state.fset.Position(ts.Pos()).Offset
				return cursor{offset: pos, endOffset: pos}, nil
			}
		}
	}
	return cursor{}, fmt.Errorf("declaration not found: %s", name)
}

func evalGotoFunc(state *applyState, rest string) (cursor, error) {
	// Parse receiver and name
	var recv string
	var funcName string

	if strings.HasPrefix(rest, "(") {
		closeParen := strings.Index(rest, ")")
		if closeParen < 0 {
			return cursor{}, fmt.Errorf("invalid func receiver: %q", rest)
		}
		recv = rest[:closeParen+1]
		funcName = strings.TrimSpace(rest[closeParen+1:])
	} else {
		funcName = rest
	}

	for _, decl := range state.astFile.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Name.Name != funcName {
			continue
		}
		if recv != "" {
			if fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			// Use the original text to match the receiver pattern
			recvStart := state.fset.Position(fd.Recv.Pos()).Offset
			recvEnd := state.fset.Position(fd.Recv.End()).Offset
			recvText := strings.TrimSpace(state.original[recvStart:recvEnd])
			if !strings.Contains(recvText, recv[1:len(recv)-1]) {
				continue
			}
		} else if fd.Recv != nil && len(fd.Recv.List) > 0 {
			continue
		}

		pos := state.fset.Position(fd.Pos()).Offset
		return cursor{offset: pos, endOffset: pos}, nil
	}

	return cursor{}, fmt.Errorf("function not found: %s", funcName)
}

func evalGotoBrace(state *applyState, opening bool) (cursor, error) {
	// Find the current declaration to locate its braces
	node := state.cursorNode()
	if node == nil {
		// Walk the AST to find declarations
		return cursor{}, fmt.Errorf("no current declaration context for brace navigation")
	}

	var lbrace, rbrace token.Pos
	switch n := node.(type) {
	case *ast.TypeSpec:
		switch t := n.Type.(type) {
		case *ast.StructType:
			if t.Fields == nil {
				return cursor{}, fmt.Errorf("struct has no fields")
			}
			lbrace = t.Fields.Opening
			rbrace = t.Fields.Closing
		case *ast.InterfaceType:
			if t.Methods == nil {
				return cursor{}, fmt.Errorf("interface has no methods")
			}
			lbrace = t.Methods.Opening
			rbrace = t.Methods.Closing
		default:
			return cursor{}, fmt.Errorf("current type is not a struct or interface")
		}
	case *ast.FuncDecl:
		if n.Body == nil {
			return cursor{}, fmt.Errorf("function has no body")
		}
		lbrace = n.Body.Lbrace
		rbrace = n.Body.Rbrace
	default:
		return cursor{}, fmt.Errorf("current node has no braces")
	}

	if !lbrace.IsValid() || !rbrace.IsValid() {
		return cursor{}, fmt.Errorf("braces not found")
	}

	if opening {
		pos := state.fset.Position(lbrace).Offset
		return cursor{offset: pos, endOffset: pos + 1}, nil
	}
	pos := state.fset.Position(rbrace).Offset
	return cursor{offset: pos, endOffset: pos + 1}, nil
}

func evalGotoField(state *applyState, fieldName string) (cursor, error) {
	node := state.cursorNode()
	ts, ok := node.(*ast.TypeSpec)
	if !ok {
		return cursor{}, fmt.Errorf("current node is not a struct")
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return cursor{}, fmt.Errorf("current node is not a struct")
	}
	if st.Fields == nil {
		return cursor{}, fmt.Errorf("struct has no fields")
	}

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				pos := state.fset.Position(name.Pos()).Offset
				return cursor{offset: pos, endOffset: pos}, nil
			}
		}
	}
	return cursor{}, fmt.Errorf("field %q not found in struct", fieldName)
}

// cursorNode returns the AST node at the current cursor position.
func (s *applyState) cursorNode() ast.Node {
	pos := token.Pos(s.fset.File(s.astFile.Pos()).Pos(s.cursor.offset))
	for _, decl := range s.astFile.Decls {
		if decl.Pos() <= pos && pos <= decl.End() {
			return s.findNodeAt(decl, pos)
		}
	}
	return s.astFile
}

func (s *applyState) findNodeAt(node ast.Node, pos token.Pos) ast.Node {
	if node == nil {
		return nil
	}
	if node.Pos() > pos || node.End() < pos {
		return nil
	}

	// try to find a more specific child node
	switch n := node.(type) {
	case *ast.GenDecl:
		for _, spec := range n.Specs {
			if child := s.findNodeAt(spec, pos); child != nil {
				return child
			}
		}
	case *ast.TypeSpec:
		if child := s.findNodeAt(n.Type, pos); child != nil {
			return child
		}
	case *ast.FuncDecl:
		return n
	case *ast.StructType:
		return node
	}
	return node
}

// evalMatch finds text within the current scope and returns a cursor.
func evalMatch(state *applyState, searchText string, forReplace bool) (cursor, error) {
	// Determine search scope
	scopeStart := 0
	scopeEnd := len(state.original)
	node := state.cursorNode()

	if node != nil && node != state.astFile {
		scopeStart = state.fset.Position(node.Pos()).Offset
		scopeEnd = state.fset.Position(node.End()).Offset
	}

	searchIn := state.original[scopeStart:scopeEnd]
	idx := strings.Index(searchIn, searchText)
	if idx < 0 {
		return cursor{}, fmt.Errorf("text not found in scope: %q", searchText)
	}

	offset := scopeStart + idx
	endOffset := offset + len(searchText)

	return cursor{
		offset:    offset,
		endOffset: endOffset,
		isReplace: forReplace,
	}, nil
}
