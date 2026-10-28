# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.3.2, 2026-10-28

- Updated Semgrep in the development image to 1.182.1 and refreshed its locked dependencies. certifi 2026.10.23 drops the emSign root CAs whose intermediate keys were compromised (GHSA-6q48-6c7w-9gw2).
- Dependabot no longer opens pull requests for `mcp` or the OpenTelemetry packages in `requirements-semgrep.txt`. Semgrep pins their versions itself, so bumping one of them on its own cannot install.
- Updated kin-openapi, which the code generator uses, from 0.142.0 to 0.149.0. The newer version fixes several panics on malformed specs reported in GitHub security advisories.
- The security scan merges its two gitleaks runs into one SARIF run. GitHub code scanning rejects a report with two runs from the same tool, so no scan results had been uploaded.

## v0.3.1, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.

## v0.3.0, 2026-10-22

- Added the account lookup endpoint `GET /api/v1/account` with `Client.GetCredentialAccountApiV1AccountGet`, `ClientWithResponses.GetCredentialAccountApiV1AccountGetWithResponse`, `NewGetCredentialAccountApiV1AccountGetRequest` and `ParseGetCredentialAccountApiV1AccountGetResponse`. It returns the account and workspace of the calling credential.
- Added the types `CredentialAccount` (with `AccountId` and `WorkspaceId`), `GetCredentialAccountApiV1AccountGetParams` (with `UnstructuredApiKey`) and `GetCredentialAccountApiV1AccountGetResponse`.
- Added `GetCredentialAccountApiV1AccountGet` to `ClientInterface` and `GetCredentialAccountApiV1AccountGetWithResponse` to `ClientWithResponsesInterface`. Code that implements either interface, such as a mock, must add the new method. The `Client` and `ClientWithResponses` types already have it.
- The partition spec moved from 1.5.132 to 1.5.137. It adds no change to the generated `partition` client.

## v0.2.0, 2026-10-21

- Removed the OAuth fields `ClientId`, `ClientSecret`, `Scope`, `Token` and `TokenUrl` from `TeradataSourceConnectorConfig`, `TeradataSourceConnectorConfigInput`, `TeradataDestinationConnectorConfig` and `TeradataDestinationConnectorConfigInput`. The upstream spec no longer has them.
- Added `UseSignedInUser` to the four Teradata connector config types, and `SkipEmptyTextElements` to `TeradataDestinationConnectorConfigInput`.
- Added `Key` to `UpdateSourceConnector` and `UpdateDestinationConnector`. The docs for `CreateSourceConnector.Key` and `CreateDestinationConnector.Key` now say the key can be edited later through `PUT`.
- Added the `Split` parameter to `GetAccountConsumedApiV1AccountsAccountIdConsumedGetParams`, with the type `GetAccountConsumedApiV1AccountsAccountIdConsumedGetParamsSplit` and the value `GetAccountConsumedApiV1AccountsAccountIdConsumedGetParamsSplitStrategy`.
- Added the `StrategyQuantity` type, and `Strategies` on `AccountConsumed` and `DayQuantity`. Added `AccountConsumed.StrategiesCountedFrom`.
- Raised the minimum Go version to 1.26.10 in `go.mod`, the dev image and the setup reference, because Go 1.26.9 has the `crypto/x509` flaw GO-2026-6809.

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
