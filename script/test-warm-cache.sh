#!/usr/bin/env bash
# Real CLI regression: a freshly installed core must seed the pre-link cache.
set -euo pipefail
xgo_bin="$(realpath "$1")"
scratch_dir="$(mktemp -d)"
cleanup() {
  result=$?
  if [ "$result" -ne 0 ]; then tail -80 "$scratch_dir"/project/*.log; fi
  rm -rf "$scratch_dir"
  exit "$result"
}
trap cleanup EXIT
export GOENV=off GOCACHEPROG="" GOTOOLCHAIN=local GOFLAGS=-trimpath
export XGO_BUILD_CACHE_DIR="$scratch_dir/cache"
xgo=("$xgo_bin" test --xgo-home "$scratch_dir/home")
mkdir -p "$scratch_dir/project" "$scratch_dir/home"
cd "$scratch_dir/project"
printf 'module example.com/warmcache\n\ngo 1.18\n' > go.mod
cat > answer_test.go <<'GO'
package warmcache
import "testing"
func answer() int { return 42 }
func TestAnswer(t *testing.T) { if answer()!=42 { t.Fatal("source change observed") } }
GO
"${xgo[@]}" --build-flag=-x . > cold.log 2>&1
"${xgo[@]}" --build-flag=-x . > warm.log 2>&1
grep -F '(cached)' warm.log
if grep -E '/link(\.exe)? -' warm.log; then
  echo 'warm cached test unexpectedly relinked' >&2
  exit 1
fi
# Explicit -a must still force compilation/linking.
"${xgo[@]}" -a --build-flag=-x . > forced.log 2>&1
grep -E '/link(\.exe)? -' forced.log > /dev/null
python3 - <<'PY'
from pathlib import Path
p=Path('answer_test.go')
p.write_text(p.read_text().replace('return 42', 'return 43'))
PY
if "${xgo[@]}" . > changed.log 2>&1; then
  echo 'source change incorrectly reused cached test result' >&2
  exit 1
fi
grep -F 'source change observed' changed.log
echo 'PASS: warm link avoided; explicit rebuild and source invalidation preserved'
