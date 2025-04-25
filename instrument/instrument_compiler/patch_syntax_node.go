package instrument_compiler

import (
	"os"
	"strings"

	"github.com/xhd2015/xgo/instrument/patch"
	"github.com/xhd2015/xgo/support/goinfo"
)

var SyntaxNodesFile = patch.FilePath{"src", "cmd", "compile", "internal", "syntax", "xgo_nodes.go"}

func PatchSyntaxNode(goroot string, goVersion *goinfo.GoVersion) error {
	if goVersion.Major > 1 || goVersion.Minor >= goinfo.GO_VERSION_22 {
		return nil
	}
	var fragments []string

	if goVersion.Major == 1 {
		if goVersion.Minor <= goinfo.GO_VERSION_21 {
			fragments = append(fragments, NodesGen)
		}
		if goVersion.Minor <= goinfo.GO_VERSION_17 {
			fragments = append(fragments, Nodes_Inspect_117)
		}
	}
	if len(fragments) == 0 {
		return nil
	}
	file := SyntaxNodesFile.JoinPrefix(goroot)
	return os.WriteFile(file, []byte("package syntax\n"+strings.Join(fragments, "\n")), 0755)
}
