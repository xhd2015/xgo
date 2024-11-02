package coverage

const help = `
Xgo tool coverage is a tool to view and merge go coverage profiles, just like an extension to go tool cover.

Usage:
    xgo tool coverage <cmd> [arguments]

The commands are:
    serve       serve profiles
    load        load profiles
    merge       merge coverage profiles
    compact     compact profile
    help        show help message

Global options:
  --project-dir DIR         the project dir

Options for merge:
    -o <file>               output to file instead of stdout
    --exclude-prefix <pkg>  exclude coverage of a specific package and sub packages

Options for serve & load:
    --diff-with REF         the base branch to diff with when displaying coverage
    --build-arg ARG         build args
    --port PORT             listening port  
    --exclude FILE          exclude FILE
    --include FILE          include FILE

Examples:
    # merge multiple files into one
    $ xgo tool coverage merge -o cover.a cover-a.out cover-b.out
    # load all
    $ xgo tool coverage load

See https://github.com/xhd2015/xgo for documentation.

`
