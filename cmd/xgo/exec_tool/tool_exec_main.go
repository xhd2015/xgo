package exec_tool

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/xgo/cmd/xgo/exec_tool/inject"
)

const ToolExecMainHelp = `
Usage: go run -toolexec="go-tool-exec <tool> --" ./hello

Tools:
 insert-trap: insert runtime.XgoTrap() to the source code

Common options:
  -v, --verbose      verbose output
  -h, --help         show help

Options for insert-trap:
  -pkg <pattern>     match package name
  -func <pattern>    match function name
`

func ToolExecMain(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: go-tool-exec <tool-name> <tool-args>")
	}

	tool := args[0]
	switch tool {
	case "insert-trap":
		return handleInsertTrap(args[1:])
	case "help", "-h", "--help":
		fmt.Print(strings.TrimPrefix(ToolExecMainHelp, "\n"))
		return nil
	default:
		return fmt.Errorf("unknown tool: %s", tool)
	}
}

// insert-trap: runtime.XgoTrap()
func handleInsertTrap(args []string) error {
	n := len(args)

	var pkgs []string
	var funcs []string
	var foundDashDash bool
	var verbose bool

	var remainArgs []string
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "--" {
			remainArgs = args[i+1:]
			foundDashDash = true
			break
		}
		if arg == "--pkg" {
			if i+1 >= n || args[i+1] == "" {
				return fmt.Errorf("%s missing arg", arg)
			}
			pkgs = append(pkgs, args[i+1])
			i++
			continue
		} else if strings.HasPrefix(arg, "--pkg=") {
			pkg := arg[len("--pkg="):]
			if pkg == "" {
				return fmt.Errorf("%s missing arg", arg)
			}
			pkgs = append(pkgs, pkg)
			continue
		}
		if arg == "--func" {
			if i+1 >= n {
				return fmt.Errorf("%s missing arg", arg)
			}
			funcs = append(funcs, args[i+1])
			i++
		} else if strings.HasPrefix(arg, "--func=") {
			fn := arg[len("--func="):]
			if fn == "" {
				return fmt.Errorf("%s missing arg", arg)
			}
			funcs = append(funcs, fn)
		}
		switch arg {
		case "-v", "--verbose":
			verbose = true
		case "-h", "--help":
			fmt.Print(strings.TrimPrefix(ToolExecMainHelp, "\n"))
			return nil
		default:
			return fmt.Errorf("unknown flag: %s", arg)
		}
	}
	if !foundDashDash {
		return errors.New("requires -- to separate tool args: go-tool-exec insert-trap --pkg <match-pattern> --func <match-pattern> -- <tool> <args>")
	}
	if len(remainArgs) == 0 {
		return errors.New("requires tool args")
	}

	if len(funcs) > 0 {
		return fmt.Errorf("--func not supported yet")
	}

	cmd := remainArgs[0]
	cmdArgs := remainArgs[1:]
	if !isCompileCommand(cmd) {
		// invoke the process as is
		runCommandExit(cmd, cmdArgs)
		return nil
	}
	if hasFlag(cmdArgs, "-V") {
		runCommandExit(cmd, cmdArgs)
		return nil
	}

	compilePkgPath := findArgAfterFlag(args, "-p")
	if compilePkgPath == "" {
		return fmt.Errorf("compile missing -p package")
	}

	if isSystemPkg(compilePkgPath) {
		runCommandExit(cmd, cmdArgs)
		return nil
	}

	if len(pkgs) > 0 && !matchAnyPkg(pkgs, compilePkgPath) {
		// let go as is
		runCommandExit(cmd, cmdArgs)
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "%s %v\n", cmd, cmdArgs)
	}

	fileIndex := findFileStartIndex(cmdArgs)
	files := cmdArgs[fileIndex:]
	if len(files) == 0 {
		// nothing to compile
		runCommandExit(cmd, cmdArgs)
		return nil
	}

	replacedFiles, err := replaceWithTmpFiles(files, verbose)
	if err != nil {
		return fmt.Errorf("failed to replace with tmp files: %w", err)
	}

	replacedCmdArgs := make([]string, len(cmdArgs))
	copy(replacedCmdArgs[:fileIndex], cmdArgs)
	for i, file := range replacedFiles {
		replacedCmdArgs[fileIndex+i] = file
	}

	runCommandExit(cmd, replacedCmdArgs)
	return nil
}

func isSystemPkg(pkg string) bool {
	if pkg == "runtime" || strings.HasPrefix(pkg, "runtime/") {
		return true
	}
	var numSlash int
	for _, c := range pkg {
		if c == '/' {
			numSlash++
			// possibly a system pkg
			if numSlash > 3 {
				return false
			}
		}
	}
	return true
}

func matchAnyPkg(pkgs []string, pkg string) bool {
	for _, p := range pkgs {
		if p == pkg {
			return true
		}
	}
	return false
}

func findFileStartIndex(args []string) int {
	n := len(args)
	for i := n - 1; i >= 0; i-- {
		if strings.HasPrefix(args[i], "-") {
			return i + 1
		}
	}
	return 0
}

func tmpRoot() string {
	if runtime.GOOS != "windows" {
		tmp := "/tmp"
		if stat, err := os.Stat(tmp); err == nil && stat.IsDir() {
			return tmp
		}
	}
	return os.TempDir()
}

// parse the ast and replace with tmp files
func replaceWithTmpFiles(files []string, verbose bool) ([]string, error) {
	tmp := filepath.Join(tmpRoot(), "runtime-xgo-trap")
	err := os.MkdirAll(tmp, 0755)
	if err != nil {
		return nil, err
	}

	var dirs []string

	replacedFiles := make([]string, len(files))
	for i, file := range files {
		if !strings.HasSuffix(file, ".go") || strings.HasSuffix(file, "_test.go") {
			replacedFiles[i] = file
			continue
		}
		absFile, err := filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		dir := filepath.Dir(absFile)
		var found bool
		for _, d := range dirs {
			if d == dir {
				found = true
				break
			}
		}
		tmpDir := filepath.Join(tmp, dir)
		if !found {
			dirs = append(dirs, dir)
			if verbose {
				fmt.Fprintf(os.Stderr, "tmp dir: %s\n", tmpDir)
			}
			// remove all files in the tmp dir
			err := os.RemoveAll(tmpDir)
			if err != nil {
				return nil, err
			}
			err = os.MkdirAll(tmpDir, 0755)
			if err != nil {
				return nil, err
			}
		}

		modifiedContent, modified, err := inject.InjectRuntimeTrap(file)
		if err != nil {
			return nil, fmt.Errorf("inject runtime.XgoTrap() to %s: %w", file, err)
		}
		if !modified {
			replacedFiles[i] = file
			continue
		}
		fileBase := filepath.Base(file)
		tmpFile := filepath.Join(tmpDir, fileBase)

		if verbose {
			fmt.Fprintf(os.Stderr, "replace file: %s -> %s\n", file, tmpFile)
		}
		err = os.WriteFile(tmpFile, modifiedContent, 0644)
		if err != nil {
			return nil, err
		}

		replacedFiles[i] = tmpFile
	}
	return replacedFiles, nil
}
