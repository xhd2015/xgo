package test_explorer

const help = `
xgo tool test-explorer is a tool used to test and debug go code easily.

Usage:
   xgo e [options]       
   xgo e [options] test

Alias:
   xgo e
   xgo explorer
   xgo test-explorer

By default, test-explorer opens a web UI to show all tests to be run.

If invoked with 'xgo e test', all tests are automatically executed without opening the web UI.

Options:
     --project-dir DIR         directory to project dir
     --go-command CMD          the command to execute test, default is 'xgo' when invoked via xgo, and 'go' otherwise
     --flag FLAG
     --flags FLAG              flags passed to test
     --exclude DIR             exclude a sub path from showing in the explorer UI
     --config FILE             test config file used to add persistent options, default: test.config.json.
                               if FILE is 'none', test config file is not read.
     --coverage[=true|false]   enable or disable coverage(default: true)
     --coverage-profile FILE   write coverage to FILE if not disabled
  -v,--verbose                 print verbose info
  -h,--help                    show help

Examples:
  xgo e                  open the test explorer in browser
  xgo e test             run all tests without opening the test explorer(used in CI)

See https://github.com/xhd2015/xgo/blob/master/doc/test-explorer/README.md for documentation.

`
