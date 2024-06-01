package debug

import (
	"fmt"
	"strings"

	"github.com/xhd2015/xgo/support/strutil"
)

func FormatDlvPrompt(port int) string {
	return FormatDlvPromptOptions(port, nil)
}

type FormatDlvOptions struct {
	VscodeExtra []string
}

func FormatDlvPromptOptions(port int, opts *FormatDlvOptions) string {
	// user need to set breakpoint explicitly
	msgs := []string{
		fmt.Sprintf("dlv listen on localhost:%d", port),
		fmt.Sprintf("Debug with IDEs:"),
		fmt.Sprintf("  > VSCode: add the following config to .vscode/launch.json configurations:"),
		fmt.Sprintf("%s", strutil.IndentLines(FormatVscodeRemoteConfig(port), "    ")),
	}
	if opts != nil {
		msgs = append(msgs, opts.VscodeExtra...)
	}
	msgs = append(msgs, []string{
		fmt.Sprintf("  > GoLand: click Add Configuration > Go Remote > localhost:%d", port),
		fmt.Sprintf("  > Terminal: dlv connect localhost:%d", port),
	}...)

	return strings.Join(msgs, "\n")
}

func FormatVscodeRemoteConfig(port int) string {
	return fmt.Sprintf(`{
	"configurations": [
		{
			"name": "Debug dlv localhost:%d",
			"type": "go",
			"debugAdapter": "dlv-dap",
			"request": "attach",
			"mode": "remote",
			"port": %d,
			"host": "127.0.0.1",
			"cwd":"./"
		}
	}
}`, port, port)
}
