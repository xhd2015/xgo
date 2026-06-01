#!/bin/bash
set -euo pipefail

# Reproduce the CI error in a podman container
# Usage:
#   bash script/reproduce-ci/reproduce.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
XGO_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
GO_VERSION="${GO_VERSION:-1.25.10}"
UBUNTU_IMAGE="${UBUNTU_IMAGE:-ubuntu:25.10}"
CONTAINER_NAME="xgo-ci-reproduce"

echo "=== XGO CI Reproduction ==="
echo "Go version: go${GO_VERSION}"
echo "xgo dir: ${XGO_DIR}"
echo "Ubuntu image: ${UBUNTU_IMAGE}"
echo ""

# Pre-download Go tarball to script dir (shared with podman VM)
GO_TARBALL="${SCRIPT_DIR}/go.tar.gz"

if [ ! -f "${GO_TARBALL}" ]; then
    echo "Downloading Go ${GO_VERSION}..."
    ARCH=$(uname -m)
    if [ "$ARCH" = "arm64" ]; then GOARCH="arm64"; else GOARCH="amd64"; fi
    curl -Lo "${GO_TARBALL}" "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz"
    echo "Downloaded: $(ls -lh ${GO_TARBALL} | awk '{print $5}')"
else
    echo "Using cached Go tarball: ${GO_TARBALL} ($(ls -lh ${GO_TARBALL} | awk '{print $5}'))"
fi

# Remove any existing container
podman rm -f "${CONTAINER_NAME}" 2>/dev/null || true

# Run container with xgo source and Go tarball mounted
podman run --rm \
    --name "${CONTAINER_NAME}" \
    -v "${XGO_DIR}:/xgo" \
    -v "${GO_TARBALL}:/tmp/go.tar.gz:ro" \
    -e "GO_VERSION=${GO_VERSION}" \
    -e "ALL_PROXY=socks5h://host.containers.internal:1080" \
    -e "NO_PROXY=localhost,127.0.0.1,*.local" \
    "${UBUNTU_IMAGE}" \
    bash "/xgo/script/reproduce-ci/startup.sh"

EXIT_CODE=$?

echo ""
echo "========================================="
echo "=== Container exit code: ${EXIT_CODE} ==="
echo "========================================="
exit ${EXIT_CODE}
