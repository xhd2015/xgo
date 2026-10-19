#!/usr/bin/env bash
# Same-workload Linux comparison. Run patch correctness suites separately.
# Optional XGO_BENCH_EXPECTED_INSTRUMENTATION_SKIPS contains newline-separated
# Package:Test names that must pass natively and skip under both xgo binaries.
# Declare only source-verified instrumentation-sensitive assertions; all other
# test names and outcomes must match exactly. Skip output stays in the report.
set -euo pipefail
if [ "$#" -lt 3 ]; then
  echo 'usage: bash script/bench-warm.sh PROJECT BASELINE_XGO CANDIDATE_XGO [go test flags/packages...]' >&2
  exit 2
fi
project_dir="$(cd "$1" && pwd)"
baseline_xgo="$(realpath "$2")"
candidate_xgo="$(realpath "$3")"
shift 3
native_go="$(command -v go)"
native_goroot="$(go env GOROOT)"
bench_dir="$(mktemp -d "${TMPDIR:-/tmp}/xgo-bench.XXXXXXXX")"
export GOENV=off GOCACHEPROG="" GOTOOLCHAIN=local
export PATH="${native_goroot}/bin:${PATH}" GOROOT="$native_goroot"
export GOTMPDIR="$bench_dir/go-tmp" TMPDIR="$bench_dir/test-tmp"
mkdir -p "$GOTMPDIR" "$TMPDIR"
cd "$project_dir"
go version
git rev-parse HEAD
"$baseline_xgo" version
"$candidate_xgo" version
echo "Benchmark output: $bench_dir"
go mod download
if [ "$#" -eq 0 ]; then set -- ./...; fi
for engine in native baseline candidate; do
  export GOCACHE="$bench_dir/$engine-cache"
  export XGO_BUILD_CACHE_DIR="$bench_dir/$engine-instrumented-cache"
  case "$engine" in
    native) command=("$native_go" test) ;;
    baseline) command=("$baseline_xgo" test) ;;
    candidate) command=("$candidate_xgo" test) ;;
  esac
  for state in cold warm1 warm2 warm3; do
    log="$bench_dir/$engine-$state.json"
    if ! /usr/bin/time -f '{"wall":%e,"user":%U,"system":%S,"max_process_rss_kb":%M}' \
      -o "$bench_dir/$engine-$state.time.json" \
      "${command[@]}" -json -short -timeout=15m -p=8 "$@" > "$log" 2>&1; then
      tail -100 "$log"
      echo "Failed: $engine $state; complete logs: $bench_dir" >&2
      exit 1
    fi
    echo "$engine $state $(cat "$bench_dir/$engine-$state.time.json")"
  done
done
python3 - "$bench_dir" <<'PY'
import collections, json, os, pathlib, statistics, sys
root = pathlib.Path(sys.argv[1])
rows = []
signatures = {}
differences = []
expected_skips = set(filter(None, os.environ.get('XGO_BENCH_EXPECTED_INSTRUMENTATION_SKIPS', '').splitlines()))
instrumentation_skips = []
for engine in ['native', 'baseline', 'candidate']:
    for state in ['cold', 'warm1', 'warm2', 'warm3']:
        counts = collections.Counter()
        tests = {}
        skip_output = []
        for line in (root/f'{engine}-{state}.json').open():
            try: event = json.loads(line)
            except json.JSONDecodeError: continue
            action = event.get('Action')
            name = event.get('Package', '')+':'+event.get('Test', '')
            if name in expected_skips and action == 'output':
                skip_output.append(event.get('Output', ''))
            if action in ['pass', 'skip', 'fail']:
                counts[('tests_' if event.get('Test') else 'packages_')+action] += 1
                if event.get('Test'):
                    tests[event['Package']+':'+event['Test']] = action
            if event.get('Output', '').startswith('ok') and '(cached)' in event['Output']:
                counts['cached_packages'] += 1
        if not tests or counts['tests_fail']:
            raise SystemExit(f'{engine}/{state}: no executed test events or failing tests')
        signature = {name: result for name, result in tests.items()}
        if not signatures:
            for name in expected_skips:
                if signature.get(name) != 'pass':
                    raise SystemExit(f'Expected instrumentation skip must pass under native Go: {name}')
        if signatures:
            reference = dict(signatures['native/cold'])
            if engine != 'native':
                reference.update({name: 'skip' for name in expected_skips})
                instrumentation_skips.append(dict(run=engine+'/'+state, names=sorted(expected_skips), output=skip_output))
        else:
            reference = signature
        if signature != reference:
            differences.append(dict(run=engine+'/'+state,
                missing=sorted(reference.keys()-signature.keys()),
                extra=sorted(signature.keys()-reference.keys()),
                changed={name: [reference[name], signature[name]] for name in reference.keys() & signature.keys() if reference[name] != signature[name]}))
        signatures[engine+'/'+state] = signature
        row = dict(engine=engine, state=state, **json.loads((root/f'{engine}-{state}.time.json').read_text()), **counts)
        rows.append(row)
median = {e: statistics.median(r['wall'] for r in rows if r['engine']==e and r['state'].startswith('warm')) for e in ['native','baseline','candidate']}
result = dict(rows=rows, warm_median_seconds=median, candidate_native_warm_ratio=median['candidate']/median['native'], target_met=median['candidate'] <= 10*median['native'], test_outcomes_match_protocol=not differences, expected_instrumentation_skips=instrumentation_skips, differences=differences)
(root/'results.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result, indent=2))
if differences:
    raise SystemExit('Test names or pass/skip results differ; inspect differences above. Timing remains provisional.')
PY
