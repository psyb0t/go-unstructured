#!/bin/bash
# Rewrites a downloaded Unstructured spec into a form oapi-codegen can generate by
# running it through a jq filter. The rewrite rules and the reasons for them
# are in the filter files next to this script.
#
# Usage: fix-spec.sh <filter.jq> <spec-in> <spec-out>
# The spec names are file names inside api/.
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

on_error() {
	local exit_code="$1"
	local line="$2"

	printf '{"level":"ERROR","file":"%s","line":%d,"msg":"command failed","exit_code":%d}\n' \
		"${BASH_SOURCE[1]}" "$line" "$exit_code" >&2
}

log() {
	local level="$1"
	shift

	jq -nc \
		--arg time "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
		--arg level "$level" \
		--arg file "${BASH_SOURCE[1]##*/}" \
		--arg line "${BASH_LINENO[0]}" \
		--arg func "${FUNCNAME[1]:-main}" \
		--arg msg "$*" \
		'{time:$time,level:$level,file:$file,line:($line|tonumber),func:$func,msg:$msg}' >&2
}

main() {
	if [ "$#" -ne 3 ]; then
		log ERROR "usage: fix-spec.sh <filter.jq> <spec-in> <spec-out>"
		exit 1
	fi

	local filter="${SCRIPT_DIR}/$1"
	local spec_in="${SPEC_DIR}/$2"
	local spec_out="${SPEC_DIR}/$3"
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${3}.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	jq -f "$filter" "$spec_in" >"$tmp"

	chmod 0644 "$tmp"
	mv -- "$tmp" "$spec_out"
	log INFO "rewrote spec filter=${1} out=${3}"
}

main "$@"
