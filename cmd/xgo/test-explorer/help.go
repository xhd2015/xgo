package test_explorer

const help = `
xgo tool test-explorer is a tool used to test and debug go code easily.

Alias:
   xgo e
   xgo explorer
   xgo test-explorer

Usage:
   xgo tool test-explorer [options]

By default, test-explorer open a web UI to show all tests to be run.

Options:
     --project-dir DIR         directory to project dir
     --go-command CMD          the command to execute test, default is 'xgo' when invoked via xgo, and 'go' otherwise
     --flag FLAG
     --flags FLAG              flags passed to test
     --exclude DIR             exclude a sub path from showing in the explorer UI
     --config FILE             test config file used to add persistent options, default: test.config.json.
                               if FILE is 'none', test config file is not read.
  -h,--help                    show help

Examples:
  xgo e                  open the test explorer in browser

See https://github.com/xhd2015/xgo/blob/master/doc/test-explorer/README.md for documentation.

`
