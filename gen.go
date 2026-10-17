// Package unstructured is a Go client for the Unstructured Platform API.
// The Partition API client is in the partition subpackage.
package unstructured

//go:generate bash scripts/fetch-spec.sh https://platform.unstructuredapp.io/openapi.json openapi.json
//go:generate bash scripts/fix-spec.sh patch-spec.jq openapi.json openapi.fixed.json
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.fixed.json
