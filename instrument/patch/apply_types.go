package patch

// PatchFile represents a parsed .xgo.patch file containing one or more <patch> blocks.
type PatchFile struct {
	Blocks []PatchBlock
}

// PatchBlock represents a single <patch name>...</patch> block.
type PatchBlock struct {
	Name     string
	Commands []Command
}

// CommandType identifies the type of a command within a patch block.
type CommandType int

const (
	CmdGoto           CommandType = iota // goto struct/func/interface/opening/closing/field
	CmdMatch                             // match <text>
	CmdFindForReplace                    // find_for_replace <text>
	CmdInsertBefore                      // insert_before <text>
	CmdInsertAfter                       // insert_after <text>
	CmdReplace                           // replace <text> (requires prior find_for_replace)
	CmdNewline                           // newline
	CmdCopyFunc                          // copy_func <source> as <target> [append to file end]
	CmdReplaceDirective                  // replace_directive <old> with <new>
	CmdInsertAfterLine                   // insert_after_line <text> (includes trailing \n in insert point)
)

// Command represents a single instruction within a <patch> block.
type Command struct {
	Type CommandType

	// For goto:
	GotoTarget string // e.g. "struct Foo", "func Bar", "func (t *T) Baz", "opening {", "closing }", "field Name"

	// For match / find_for_replace / replace_directive (old text):
	SearchText string

	// For insert_before / insert_after / replace / replace_directive (new text):
	EditText string

	// For copy_func / replace_directive:
	CopySource string // source function name
	CopyTarget string // target function name or replace_directive new text
}

// String returns a human-readable representation of the command.
func (c Command) String() string {
	switch c.Type {
	case CmdGoto:
		return "goto " + c.GotoTarget
	case CmdMatch:
		return "match " + c.SearchText
	case CmdFindForReplace:
		return "find_for_replace " + c.SearchText
	case CmdInsertBefore:
		return "insert_before " + c.EditText
	case CmdInsertAfter:
		return "insert_after " + c.EditText
	case CmdInsertAfterLine:
		return "insert_after_line " + c.EditText
	case CmdReplace:
		return "replace " + c.EditText
	case CmdNewline:
		return "newline"
	case CmdCopyFunc:
		return "copy_func " + c.CopySource + " as " + c.CopyTarget + " append to file end"
	case CmdReplaceDirective:
		return "replace_directive " + c.SearchText + " with " + c.CopyTarget
	default:
		return "unknown"
	}
}
