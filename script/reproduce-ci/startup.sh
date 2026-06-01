#!/bin/bash
set -euo pipefail

# Setup script that runs inside the podman container
# Go tarball is expected to be pre-downloaded and mounted at /tmp/go.tar.gz

GO_VERSION="${GO_VERSION:-1.25.10}"

echo "========================================="
echo "=== Step 1: Installing Go ${GO_VERSION} ==="
echo "========================================="
tar -C /usr/local -xzf /tmp/go.tar.gz

export GOROOT=/usr/local/go
export GOPATH=/root/go
export PATH=/usr/local/go/bin:/root/go/bin:$PATH

echo ""
go version

echo ""
echo "========================================="
echo "=== Step 2: Running run-test ==="
echo "========================================="
echo "Command: go run ./script/run-test/ --no-git --install-xgo --reset-instrument --log-debug -v -short --with-goroot /usr/local/go"
echo ""

cd /xgo
go run ./script/run-test/ \
    --no-git \
    --install-xgo \
    --reset-instrument \
    --log-debug \
    -v \
    -short \
    --with-goroot /usr/local/go

EXIT_CODE=$?

echo ""
echo "========================================="
echo "=== Result ==="
echo "========================================="
if [ ${EXIT_CODE} -eq 0 ]; then
    echo "SUCCESS: All tests passed"
else
    echo "FAILURE: Tests failed with exit code ${EXIT_CODE}"
fi
exit ${EXIT_CODE}
