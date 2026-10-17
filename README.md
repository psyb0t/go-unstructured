# go-unstructured

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-unstructured.svg)](https://pkg.go.dev/github.com/psyb0t/go-unstructured)
[![CI](https://github.com/psyb0t/go-unstructured/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-unstructured/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-unstructured/badges/version.svg)](https://github.com/psyb0t/go-unstructured/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-unstructured/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-unstructured/badges/importers.svg)](https://github.com/psyb0t/go-unstructured/blob/badges/importers.md)

Unofficial Go client for the [Unstructured Platform API](https://platform.unstructuredapp.io/openapi.json) and [Partition API](https://api.unstructuredapp.io/general/openapi.json), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from Unstructured's public OpenAPI documents. I am not affiliated with Unstructured.

Package `unstructured` in the repository root covers sources, destinations, workflows, jobs, templates, notifications and usage. Package `partition` handles document partitioning. Both `client.gen.go` files are generated. Regenerate them rather than editing them by hand.

## Install

```sh
go get github.com/psyb0t/go-unstructured
```

## Usage

Every operation in each spec has a method on its package's `ClientWithResponses`. JSON responses decode into `JSON200` and the other status fields. Check the status and decoded field before using a response; an HTTP error does not necessarily return a Go error.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-unstructured"
)

func main() {
	auth := unstructured.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"unstructured-api-key",
				os.Getenv("UNSTRUCTURED_API_KEY"),
			)

			return nil
		},
	)

	client, err := unstructured.NewClientWithResponses(
		"https://platform.unstructuredapp.io",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.ListSourcesWithResponse(
		context.Background(),
		&unstructured.ListSourcesParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles. Import `github.com/psyb0t/go-unstructured/partition` for the Partition API. For uploads, use `PartitionWithBodyWithResponse` with a multipart body and the multipart writer's content type. Parsed JSON output is a slice of `partition.Element` values.

Method names come from spec operation IDs. Platform notification operations have long route-based names; source, destination, workflow and job methods use names such as `ListSourcesWithResponse` and `GetJobWithResponse`.

The full API, with the spec's descriptions as doc comments, is in the [Go reference](https://pkg.go.dev/github.com/psyb0t/go-unstructured) and in `go doc github.com/psyb0t/go-unstructured` (or `.../partition`) from inside your module.

## Auth

Both APIs send the API key in the `unstructured-api-key` header. The request editor above reads it from `UNSTRUCTURED_API_KEY`. The Platform spec declares optional header parameters; the Partition spec declares an `ApiKeyAuth` header scheme and an anonymous alternative. Set the header explicitly for authenticated requests. Never log the key or a request containing it.

| API | Base URL |
| --- | --- |
| Platform | `https://platform.unstructuredapp.io` |
| Partition | `https://api.unstructuredapp.io` |

## Regenerating

The specs are not committed. `make generate` runs `go generate ./...` and downloads both documents into the gitignored `api/` directory:

```sh
make generate
```

- `scripts/fetch-spec.sh` downloads the Platform spec to `api/openapi.json` and the Partition spec to `api/partition.json`.
- `scripts/fix-spec.sh` applies `scripts/patch-spec.jq`, converting nullable `anyOf` unions to `nullable`, collapsing singleton unions, converting `const` to `enum` and schema `examples` to `example`, then writing OpenAPI 3.0.3 documents for oapi-codegen.
- The patch also reduces redundant UUID-or-string unions to `string`. This preserves accepted values and avoids generated path and query parameter structs whose private union fields callers cannot populate.
- Open-ended additional properties and unspecified JSON response bodies use `json.RawMessage`, preserving their JSON values and large integers without untyped deserialization.
- oapi-codegen runs with `config.yaml` and `partition/config.yaml`. Both map `number` to `float64` and generate all operations.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-unstructured) teaches a coding agent how to use both packages. Agents that read `.agents/skills/` can use it directly from a checkout. The release workflow packages it for the channels below; installation requires a published release and marketplace entry.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-unstructured@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-unstructured@psyb0t
```

### OpenClaw

The release workflow publishes the skill to ClawHub:

```bash
openclaw skills install @psyb0t/go-unstructured
```

## Development

Every Make target runs its tools in the `Dockerfile.dev` image. The host needs Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make generate  # download both specs and regenerate
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
```

## License

[MIT](LICENSE). The generated code comes from Unstructured's published API specifications. The specifications themselves are not redistributed here. Neither downloaded document declares a specification license in its `info` object.
