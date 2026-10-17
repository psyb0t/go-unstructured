// Package partition is a Go client for the Unstructured Partition API.
package partition

//go:generate bash ../scripts/fetch-spec.sh https://api.unstructuredapp.io/general/openapi.json partition.json
//go:generate bash ../scripts/fix-spec.sh patch-spec.jq partition.json partition.fixed.json
//go:generate go tool oapi-codegen --config=config.yaml ../api/partition.fixed.json
