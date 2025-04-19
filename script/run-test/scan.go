package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func list(args []string, flagJSON bool, predefinedTests []*TestConfig) {
	tests := getTests(args, predefinedTests)
	if flagJSON {
		data, err := json.Marshal(tests)
		if err != nil {
			panic(fmt.Errorf("json.Marshal: %v", err))
		}
		fmt.Println(string(data))
		return
	}
	if len(tests) == 0 {
		fmt.Fprintf(os.Stderr, "no tests\n")
		return
	}
	for _, testArg := range tests {
		logDir := testArg.Dir
		if len(testArg.Args) == 0 {
			fmt.Fprintf(os.Stdout, "./%s\n", logDir)
			continue
		}
		prefix := "."
		if logDir != "" {
			prefix = "./" + logDir
		}
		for _, arg := range testArg.Args {
			fmt.Fprintf(os.Stdout, "%s/%s\n", prefix, strings.TrimPrefix(arg, "./"))
		}
	}
}

func getTests(args []string, predefinedTests []*TestConfig) []*TestConfig {
	tests := scanTests(predefinedTests)
	if len(args) == 0 {
		err := amedTestsWithConfig(tests, true)
		if err != nil {
			panic(fmt.Errorf("amedTestsWithConfig: %v", err))
		}
		return tests
	}
	resolvedTests := resolveTests(tests, defaultTest, args)
	err := amedTestsWithConfig(resolvedTests, false)
	if err != nil {
		panic(fmt.Errorf("amedTestsWithConfig: %v", err))
	}
	return resolvedTests
}

const testConfigFile = "test-config.txt"

func amedTestsWithConfig(tests []*TestConfig, includePredefinedArgs bool) error {
	for _, test := range tests {
		if err := amedTestWithConfig(test, includePredefinedArgs); err != nil {
			return err
		}
	}
	return nil
}
func amedTestWithConfig(test *TestConfig, includePredefinedArgs bool) error {
	file := filepath.Join(test.Dir, testConfigFile)
	bytes, readErr := os.ReadFile(file)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return readErr
		}
	}
	type handle struct {
		prefix string
		f      func(content string)
	}
	var predefinedArgs []string
	handles := []handle{
		{prefix: "flags:", f: func(content string) {
			extraFlags := splitList(content)
			test.Flags = append(test.Flags, extraFlags...)
		}},
		{prefix: "args:", f: func(content string) {
			predefinedArgs = append(predefinedArgs, splitList(content)...)
		}},
		{prefix: "use-plain-go:", f: func(content string) {
			test.UsePlainGo = content == "true"
		}},
		{prefix: "use-prebuilt-xgo:", f: func(content string) {
			test.UsePrebuiltXgo = content == "true"
		}},
		{prefix: "build-only:", f: func(content string) {
			test.BuildOnly = content == "true"
		}},
		{prefix: "vendor-if-missing:", f: func(content string) {
			test.VendorIfMissing = content == "true"
		}},
		{prefix: "env:", f: func(content string) {
			test.Env = append(test.Env, splitList(content)...)
		}},
		{prefix: "go:", f: func(content string) {
			test.Go = content
		}},
	}
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for _, h := range handles {
			if strings.HasPrefix(line, h.prefix) {
				h.f(strings.TrimSpace(line[len(h.prefix):]))
				break
			}
		}
	}
	if includePredefinedArgs {
		if len(test.Args) == 0 && len(predefinedArgs) == 0 {
			test.Args = []string{"./..."}
		} else {
			test.Args = append(test.Args, predefinedArgs...)
		}
	} else if len(test.Args) == 1 && test.Args[0] == "./all" {
		if len(predefinedArgs) == 0 {
			test.Args = []string{"./..."}
		} else {
			test.Args = predefinedArgs
		}
	}
	return nil
}
func splitList(content string) []string {
	var list []string
	n := len(content)

	var buf []byte
	for i := 0; i < n; i++ {
		b := content[i]
		// "\ " -> escape a space
		if b == '\\' && (i+1 < n && content[i+1] == ' ') {
			buf = append(buf, ' ')
			i++
			continue
		}
		if b != ' ' {
			buf = append(buf, b)
			continue
		}
		if len(buf) > 0 {
			list = append(list, string(buf))
			buf = nil
		}
	}
	if len(buf) > 0 {
		list = append(list, string(buf))
	}
	return list
}

func scanTests(predefinedTests []*TestConfig) []*TestConfig {
	subTests, err := scanGoMods(filepath.Join("runtime", "test"), []string{"runtime", "test"})
	if err != nil {
		panic(fmt.Errorf("scanGoMods: %v", err))
	}
	tests := make([]*TestConfig, 0, len(subTests))
	tests = append(tests, predefinedTests...)
	for _, subTest := range subTests {
		subTestDir := strings.Join(subTest, "/")
		var found bool
		for _, test := range predefinedTests {
			if test.Dir == subTestDir {
				found = true
				break
			}
		}
		if found {
			continue
		}
		tests = append(tests, &TestConfig{Dir: subTestDir})
	}
	return tests
}

// scan all dirs that contains go.mod
// return a list of subPaths related to dir
func scanGoMods(dir string, prefix []string) ([][]string, error) {
	names, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var results [][]string
	for _, name := range names {
		if name.Name() == "go.mod" && !name.IsDir() {
			results = append(results, appendCopy(prefix, nil))
			continue
		}
	}
	for _, name := range names {
		if !name.IsDir() {
			continue
		}
		nameStr := name.Name()
		if nameStr == "vendor" || nameStr == "testdata" || strings.HasPrefix(nameStr, ".") {
			// .git, .xgo etc
			continue
		}
		subPrefix := appendCopy(prefix, []string{nameStr})

		subMods, err := scanGoMods(filepath.Join(dir, nameStr), subPrefix)
		if err != nil {
			return nil, err
		}
		results = append(results, subMods...)
	}
	return results, nil
}

func appendCopy(prefix []string, suffix []string) []string {
	list := make([]string, len(prefix)+len(suffix))
	copy(list, prefix)
	copy(list[len(prefix):], suffix)
	return list
}

// resolveTests associate args to tests
func resolveTests(tests []*TestConfig, defaultTest *TestConfig, args []string) []*TestConfig {
	if len(args) == 0 {
		return tests
	}
	var results []*TestConfig
	for _, arg := range args {
		path, test := resolveTestConfig(tests, arg)
		if test == nil {
			// fallback to default test
			test = defaultTest
			path = arg
		}

		if test == nil {
			continue
		}

		var prev *TestConfig
		for i, r := range results {
			if r.Dir == test.Dir {
				prev = results[i]
				break
			}
		}
		if prev == nil {
			clone := *test
			clone.Args = nil
			prev = &clone
			results = append(results, prev)
		}
		if path == "" {
			path = "./"
		}
		prev.Args = append(prev.Args, path)
	}
	return results
}

var xgoModule = []string{"github.com", "xhd2015", "xgo"}

func resolveTestConfig(tests []*TestConfig, fullArg string) (path string, test *TestConfig) {
	rootTest := findTest(tests, "")
	if fullArg == "" {
		return "", rootTest
	}
	parts := splitPath(fullArg)
	if len(parts) == 0 {
		return "", rootTest
	}
	var xgoPath []string
	if parts[0] == "." {
		xgoPath = parts[1:]
	} else if listEquals(parts, xgoModule) {
		xgoPath = parts[len(xgoModule):]
	} else {
		return fullArg, rootTest
	}
	if len(xgoPath) == 0 || xgoPath[0] != "runtime" {
		return fullArg, rootTest
	}

	// runtime
	runtimePath := xgoPath[1:]
	if len(runtimePath) == 0 || runtimePath[0] != "test" {
		runtimeTest := findTest(tests, "runtime")
		return strings.Join(runtimePath, "/"), runtimeTest
	}

	runtimeTestPath := runtimePath[1:]
	// runtime/test
	_, remains, test := findMostInnerTest(tests, "runtime/test", runtimeTestPath)
	modPath := "./"
	if len(remains) > 0 {
		modPath = "./" + strings.Join(remains, "/")
	}
	return modPath, test
}

func listEquals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, e := range a {
		if e != b[i] {
			return false
		}
	}
	return true
}

func findMostInnerTest(tests []*TestConfig, prefixDir string, path []string) (dirPath []string, suffixPath []string, test *TestConfig) {
	n := len(path)
	for i := n - 1; i >= 0; i-- {
		testPath := path[:i+1]
		test := findTest(tests, filepath.Join(prefixDir, filepath.Join(testPath...)))
		if test != nil {
			return testPath, path[i+1:], test
		}
	}
	test = findTest(tests, prefixDir)
	return nil, path, test
}

func findTest(tests []*TestConfig, dir string) *TestConfig {
	for _, test := range tests {
		if test.Dir == dir {
			return test
		}
	}
	return nil
}
