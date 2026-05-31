package patch

import (
	"fmt"
	"strings"
)

// ParseXgoPatch parses a .xgo.patch file content into structured blocks and commands.
func ParseXgoPatch(content string) (*PatchFile, error) {
	blocks, err := parseBlocks(content)
	if err != nil {
		return nil, err
	}
	return &PatchFile{Blocks: blocks}, nil
}

func parseBlocks(content string) ([]PatchBlock, error) {
	var blocks []PatchBlock

	for {
		openIdx := strings.Index(content, "<patch")
		if openIdx < 0 {
			break
		}

		openTagEnd := strings.Index(content[openIdx:], ">")
		if openTagEnd < 0 {
			return nil, fmt.Errorf("unclosed <patch> tag")
		}
		openTagEnd += openIdx + 1

		closeIdx := strings.Index(content[openTagEnd:], "</patch>")
		if closeIdx < 0 {
			return nil, fmt.Errorf("missing </patch> for %q", content[openIdx:openIdx+min(len(content)-openIdx, 50)])
		}
		closeIdx += openTagEnd
		closeEnd := closeIdx + len("</patch>")

		block, err := parseBlock(content[openIdx:closeEnd])
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
		content = content[closeEnd:]
	}

	return blocks, nil
}

func parseBlock(blockContent string) (PatchBlock, error) {
	openIdx := strings.Index(blockContent, "<patch")
	openTagEnd := strings.Index(blockContent[openIdx:], ">")
	openTagEnd += openIdx + 1

	tagContent := blockContent[openIdx+len("<patch") : openTagEnd-1]
	tagContent = strings.TrimSpace(tagContent)

	closeIdx := strings.Index(blockContent[openTagEnd:], "</patch>")
	closeIdx += openTagEnd
	bodyContent := blockContent[openTagEnd:closeIdx]

	block := PatchBlock{Name: tagContent}

	lines := strings.Split(bodyContent, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		cmd, err := parseLine(line)
		if err != nil {
			return PatchBlock{}, fmt.Errorf("in patch %q: %w", tagContent, err)
		}
		block.Commands = append(block.Commands, cmd)
	}

	return block, nil
}

func parseLine(line string) (Command, error) {
	cmd, rest := splitCommand(line)
	switch {
	case cmd == "goto":
		return parseGotoTarget(rest)
	case cmd == "match":
		return Command{Type: CmdMatch, SearchText: rest}, nil
	case cmd == "find_for_replace":
		return Command{Type: CmdFindForReplace, SearchText: rest}, nil
	case cmd == "insert_before":
		if rest == "" {
			return Command{}, fmt.Errorf("insert_before requires text")
		}
		return Command{Type: CmdInsertBefore, EditText: rest}, nil
	case cmd == "insert_after_line":
		if rest == "" {
			return Command{}, fmt.Errorf("insert_after_line requires text")
		}
		return Command{Type: CmdInsertAfterLine, EditText: rest}, nil
	case cmd == "insert_after":
		if rest == "" {
			return Command{}, fmt.Errorf("insert_after requires text")
		}
		return Command{Type: CmdInsertAfter, EditText: rest}, nil
	case cmd == "replace":
		if rest == "" {
			return Command{}, fmt.Errorf("replace requires text")
		}
		return Command{Type: CmdReplace, EditText: rest}, nil
	case cmd == "newline":
		return Command{Type: CmdNewline}, nil
	case cmd == "copy_func":
		return parseCopyFuncLine(rest)
	case cmd == "replace_directive":
		return parseReplaceDirectiveLine(rest)
	default:
		return Command{}, fmt.Errorf("unknown command: %q", line)
	}
}

func splitCommand(line string) (cmd string, rest string) {
	line = strings.TrimRight(line, " \t\r")
	trimmed := strings.TrimLeft(line, " \t")
	leadingLen := len(line) - len(trimmed)
	spaceIdx := strings.Index(trimmed, " ")
	if spaceIdx < 0 {
		return trimmed, ""
	}
	cmd = trimmed[:spaceIdx]
	rest = line[leadingLen+spaceIdx+1:]
	return
}

func parseGoto(line string) (Command, error) {
	target := strings.TrimSpace(line[len("goto "):])
	return parseGotoTarget(target)
}

func parseGotoTarget(target string) (Command, error) {
	if target == "opening {" || target == "closing }" {
		return Command{Type: CmdGoto, GotoTarget: target}, nil
	}

	if strings.HasPrefix(target, "field ") {
		return Command{Type: CmdGoto, GotoTarget: target}, nil
	}

	if strings.HasPrefix(target, "struct ") {
		return Command{Type: CmdGoto, GotoTarget: target}, nil
	}

	if strings.HasPrefix(target, "interface ") {
		return Command{Type: CmdGoto, GotoTarget: target}, nil
	}

	if strings.HasPrefix(target, "func ") {
		return Command{Type: CmdGoto, GotoTarget: target}, nil
	}

	return Command{}, fmt.Errorf("unknown goto target: %q", target)
}

func parseCopyFuncLine(rest string) (Command, error) {
	rest = strings.TrimSpace(rest)
	asIdx := strings.Index(rest, " as ")
	if asIdx < 0 {
		return Command{}, fmt.Errorf("copy_func requires 'as' keyword: %q", rest)
	}
	source := rest[:asIdx]
	target := strings.TrimSuffix(rest[asIdx+4:], " append to file end")
	return Command{
		Type:       CmdCopyFunc,
		CopySource: source,
		CopyTarget: target,
	}, nil
}

func parseReplaceDirectiveLine(rest string) (Command, error) {
	rest = strings.TrimSpace(rest)
	withIdx := strings.Index(rest, " with ")
	if withIdx < 0 {
		return Command{}, fmt.Errorf("replace_directive requires 'with' keyword: %q", rest)
	}
	oldText := rest[:withIdx]
	newText := rest[withIdx+len(" with "):]
	return Command{
		Type:       CmdReplaceDirective,
		SearchText: oldText,
		CopyTarget: newText,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
