#!/bin/bash
# Downloads one Unstructured OpenAPI spec into api/. The specs are not committed;
# `go generate` fetches them as its first step.
#
# Usage: fetch-spec.sh <spec-url> <file-name>
set -euo pipefail
trap 'on_error "$?" "$LINENO"' ERR

LOG_FILE="${LOG_FILE:-/tmp/$(basename "$0" .sh).log}"
exec > >(tee -a "$LOG_FILE") 2>&1

readonly DOWNLOAD_TIMEOUT_SECONDS=60
SPEC_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/api"
readonly SPEC_DIR

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
	if [ "$#" -ne 2 ]; then
		log ERROR "usage: fetch-spec.sh <spec-url> <file-name>"
		exit 1
	fi

	local spec_url="$1"
	local spec_name="$2"

	mkdir -p "$SPEC_DIR"
	local tmp
	tmp="$(mktemp "${SPEC_DIR}/.${spec_name}.XXXXXX")"
	# shellcheck disable=SC2064 # expand tmp now, the local is gone when EXIT fires
	trap "rm -f -- '${tmp}'" EXIT

	log INFO "downloading spec=${spec_name}"
	wget -q \
		-T "$DOWNLOAD_TIMEOUT_SECONDS" \
		-O "$tmp" \
		"$spec_url"

	if ! jq -e '.openapi | type == "string"' "$tmp" >/dev/null; then
		log ERROR "downloaded file is not an OpenAPI document spec=${spec_name}"
		exit 1
	fi

	chmod 0644 "$tmp"
	mv -- "$tmp" "${SPEC_DIR}/${spec_name}"
	log INFO "saved spec=${spec_name} openapi=$(jq -r '.openapi' "${SPEC_DIR}/${spec_name}") version=$(jq -r '.info.version' "${SPEC_DIR}/${spec_name}")"
}

main "$@"
