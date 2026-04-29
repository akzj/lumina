#!/usr/bin/env bash
# Run Lumina Lua integration tests (see docs/TESTING.md).
#
#   ./scripts/lua-test.sh
#   ./scripts/lua-test.sh -- -v
#   ./scripts/lua-test.sh examples/scrollview_test.lua
#   ./scripts/lua-test.sh examples/wm_test.lua -- -v
#   ./scripts/lua-test.sh scrollview          # unique substring under testdata/lua_tests/

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/pkg"

go_args=() # always set; set -u safe for "${go_args[@]}"
if [[ $# -ge 1 && "$1" == "--" ]]; then
	shift
	go_args=("$@")
	exec go test . -run TestLuaTestFramework -count=1 ${go_args[@]+"${go_args[@]}"}
fi

selector=""
if [[ $# -ge 1 ]]; then
	selector="$1"
	shift
fi
if [[ $# -ge 1 && "$1" == "--" ]]; then
	shift
	go_args=("$@")
fi

if [[ -z "$selector" ]]; then
	exec go test . -run TestLuaTestFramework -count=1 ${go_args[@]+"${go_args[@]}"}
fi

rel=""
if [[ -f "testdata/lua_tests/$selector" ]]; then
	rel="testdata/lua_tests/$selector"
elif [[ -f "$selector" ]]; then
	rel="$selector"
fi

if [[ -z "$rel" ]]; then
	matches=()
	while IFS= read -r line; do
		[[ -n "$line" ]] && matches+=("$line")
	done < <(find testdata/lua_tests -name '*_test.lua' -type f 2>/dev/null | grep -F "$selector" || true)
	if [[ ${#matches[@]} -eq 1 ]]; then
		rel="${matches[0]}"
	elif [[ ${#matches[@]} -gt 1 ]]; then
		echo "Ambiguous '$selector':" >&2
		printf '  %s\n' "${matches[@]}" >&2
		exit 1
	fi
fi

if [[ -z "$rel" ]] || [[ ! -f "$rel" ]]; then
	echo "Unknown test file or pattern: $selector" >&2
	echo "Example: ./scripts/lua-test.sh examples/scrollview_test.lua" >&2
	exit 1
fi

export LUMINA_LUA_TEST="$rel"
exec go test . -run TestLuaTestFramework -count=1 ${go_args[@]+"${go_args[@]}"}
