package partition_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/psyb0t/go-unstructured/partition"
)

//nolint:testableexamples // Requires live API credentials; compile check only.
func Example() {
	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	file, err := writer.CreateFormFile("files", "document.txt")
	if err != nil {
		panic(err)
	}

	if _, err := io.WriteString(file, "A short document."); err != nil {
		panic(err)
	}

	if err := writer.Close(); err != nil {
		panic(err)
	}

	auth := partition.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"unstructured-api-key",
				os.Getenv("UNSTRUCTURED_API_KEY"),
			)

			return nil
		},
	)

	client, err := partition.NewClientWithResponses(
		"https://api.unstructuredapp.io",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.PartitionWithBodyWithResponse(
		context.Background(),
		&partition.PartitionParams{},
		writer.FormDataContentType(),
		&body,
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
