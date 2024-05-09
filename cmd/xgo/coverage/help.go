package coverage

const help = `
Xgo tool coverage is a tool to view and merge go coverage profiles, just like an extension to go tool cover.

Usage:
    xgo tool coverage <cmd> [arguments]

The commands are:
    merge       merge coverage profiles
    help        show help message

Options for merge:
    -o <file>               output to file instead of stdout
    --exclude-prefix <pkg>  exclude coverage of a specific package and sub packages

Examples:
    xgo tool coverage merge -o cover.a cover-a.out cover-b.out     merge multiple files into one

See https://github.com/xhd2015/xgo for documentation.

`
