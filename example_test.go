package unstructured_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-unstructured"
)

//nolint:testableexamples // Requires live API credentials; compile check only.
func Example() {
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
