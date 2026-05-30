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
		line = strings.TrimSpace(line)
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
	switch {
	case strings.HasPrefix(line, "goto "):
		return parseGoto(line)
	case strings.HasPrefix(line, "match "):
		return Command{Type: CmdMatch, SearchText: strings.TrimSpace(line[len("match "):])}, nil
	case strings.HasPrefix(line, "find_for_replace "):
		return Command{Type: CmdFindForReplace, SearchText: strings.TrimSpace(line[len("find_for_replace "):])}, nil
	case strings.HasPrefix(line, "insert_before "):
		text := line[len("insert_before "):]
		if text == "" {
			return Command{}, fmt.Errorf("insert_before requires text")
		}
		return Command{Type: CmdInsertBefore, EditText: text}, nil
	case strings.HasPrefix(line, "insert_after "):
		text := line[len("insert_after "):]
		if text == "" {
			return Command{}, fmt.Errorf("insert_after requires text")
		}
		return Command{Type: CmdInsertAfter, EditText: text}, nil
	case strings.HasPrefix(line, "replace "):
		text := line[len("replace "):]
		if text == "" {
			return Command{}, fmt.Errorf("replace requires text")
		}
		return Command{Type: CmdReplace, EditText: text}, nil
	case line == "newline":
		return Command{Type: CmdNewline}, nil
	default:
		return Command{}, fmt.Errorf("unknown command: %q", line)
	}
}

func parseGoto(line string) (Command, error) {
	target := strings.TrimSpace(line[len("goto "):])

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
