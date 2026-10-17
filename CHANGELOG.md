# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.0, 2026-10-17

- Added `unstructured`, a REST client generated with oapi-codegen from the Unstructured Platform API OpenAPI 3.1 document, spec version 3.3.0, at https://platform.unstructuredapp.io/openapi.json. It covers sources, destinations, workflows, jobs, templates, notifications and usage.
- Added `partition`, a REST client generated from the Unstructured Partition API OpenAPI 3.1 document, spec version 1.5.132, at https://api.unstructuredapp.io/general/openapi.json. It covers multipart document partitioning.
- Number fields are `float64` instead of oapi-codegen's default `float32`.
- `make generate` downloads both specs, converts unsupported OpenAPI 3.1 schema forms to 3.0 forms and regenerates both clients. The specs are not committed.
- Redundant UUID-or-string unions become `string`, preserving accepted values and producing usable path and query parameters.
- Open-ended additional properties and unspecified JSON response bodies use `json.RawMessage` to preserve JSON values and large integers without untyped deserialization.
- Added the `go-unstructured` agent skill with setup and API usage references.
