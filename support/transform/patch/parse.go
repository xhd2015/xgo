package patch

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type PatchContent struct {
	Prepend *PatchLines
	Append  *PatchLines
	Replace *PatchLines
}

type PatchLines struct {
	ID    string
	Lines []string
}

func parsePatch(lines []string, node ast.Node, fset *token.FileSet) map[ast.Node]*PatchContent {
	assignedLine := make(map[int]bool)
	mapping := make(map[ast.Node]*PatchContent)
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch n := n.(type) {
		case *ast.Comment:
			return false
		case *ast.CommentGroup:
			return false
		case *ast.GenDecl:
			if n.Lparen == token.NoPos {
				// if the decl does not include a (
				// then the comment will be assigned to specific node
				return true
			}
		}
		comments := getCommentLines(lines, n, fset, assignedLine)
		if len(comments) == 0 {
			return true
		}
		p, _ := parseComments(comments)
		if p != nil {
			mapping[n] = p
		}
		return true
	})
	return mapping
}

const DOUBLE_SLASH = "//"

func getCommentLines(lines []string, node ast.Node, fset *token.FileSet, assignedLine map[int]bool) []string {
	line := fset.Position(node.Pos()).Line

	if line <= 1 || assignedLine[line-1] {
		return nil
	}

	var comments []string
	// previous line
	for i := line - 1; i > 0; i-- {
		if assignedLine[i] {
			continue
		}
		trimLine := strings.TrimSpace(lines[i-1])
		if !strings.HasPrefix(trimLine, DOUBLE_SLASH) {
			comments = make([]string, 0, line-i-1)
			for j := i + 1; j < line; j++ {
				assignedLine[j] = true
				comments = append(comments, strings.TrimSpace(lines[j-1])[len(DOUBLE_SLASH):])
			}
			break
		}
	}
	return comments
}

type cmd struct {
	command string
	lines   []string
	id      string
}

func parseComments(comments []string) (*PatchContent, error) {
	appendCmd := &cmd{
		command: "append",
	}
	prependCmd := &cmd{
		command: "prepend",
	}
	replaceCmd := &cmd{
		command: "replace",
	}
	cmds := []*cmd{appendCmd, prependCmd, replaceCmd}

	var lastCmd *cmd

	n := len(comments)
	for i := 0; i < n; i++ {
		line := comments[i]
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || trimmedLine == "..." {
			continue
		}
		if strings.HasPrefix(line, " ") {
			// start with space
			if lastCmd == nil {
				return nil, fmt.Errorf("unknown text: %s", line)
			}
			lastCmd.lines = append(lastCmd.lines, strings.TrimSpace(line))
			continue
		}
		var match bool
		for _, cmd := range cmds {
			id, content, ok, err := tryCommand(line, cmd.command)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			lastCmd = cmd
			match = true
			if id != "" {
				if cmd.id == "" {
					cmd.id = id
				} else if cmd.id != id {
					return nil, fmt.Errorf("duplicate id: %s", line)
				}
			}
			cmd.lines = append(cmd.lines, content)
			break
		}
		if !match {
			return nil, fmt.Errorf("bad instruction: %s", line)
		}
	}

	for _, cmd := range cmds {
		if len(cmd.lines) > 0 && cmd.id == "" {
			return nil, fmt.Errorf("missing id: %s", cmd.lines[0])
		}
	}

	return &PatchContent{
		Prepend: &PatchLines{ID: prependCmd.id, Lines: prependCmd.lines},
		Append:  &PatchLines{ID: appendCmd.id, Lines: appendCmd.lines},
		Replace: &PatchLines{ID: replaceCmd.id, Lines: replaceCmd.lines},
	}, nil
}

func tryCommand(line string, cmd string) (id string, content string, ok bool, err error) {
	prefix := cmd + " "
	if !strings.HasPrefix(line, prefix) {
		return
	}
	ok = true
	remain := strings.TrimSpace(line[len(prefix):])

	if strings.HasPrefix(remain, "<") {
		idx := strings.Index(remain[1:], ">")
		if idx < 0 {
			err = fmt.Errorf("missing >")
			return
		}
		id = remain[1 : idx+1]
		remain = remain[idx+2:]
	}

	content = remain
	return
}
