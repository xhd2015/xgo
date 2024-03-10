#!/usr/bin/env bash

shdir=$(cd "$(dirname "$0")" && pwd)
goroot=$(cd "$shdir/../../../../../.." && pwd)

export GOROOT=$goroot
export PATH=$goroot/bin:$PATH

env -- "$@"
