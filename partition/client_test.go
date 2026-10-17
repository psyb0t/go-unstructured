package partition_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/psyb0t/go-unstructured/partition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey        = "EXAMPLE-DO-NOT-USE"
	apiKeyHeader      = "unstructured-api-key"
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
	partitionPath     = "/general/v0/general"
	documentText      = "A short document."
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *partition.ClientWithResponses {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	auth := partition.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(apiKeyHeader, testAPIKey)

			return nil
		},
	)
	client, err := partition.NewClientWithResponses(server.URL, auth)
	require.NoError(t, err)

	return client
}

func multipartDocument(t *testing.T) (string, *bytes.Buffer) {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("files", "document.txt")
	require.NoError(t, err)
	_, err = io.WriteString(file, documentText)
	require.NoError(t, err)
	require.NoError(t, writer.WriteField("strategy", "fast"))
	require.NoError(t, writer.Close())

	return writer.FormDataContentType(), &body
}

func assertRequest(t *testing.T, r *http.Request) {
	t.Helper()

	assert.Equal(t, http.MethodPost, r.Method)
	assert.Equal(t, partitionPath, r.URL.Path)
	assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
}

func TestPartitionDocument(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r)

		reader, err := r.MultipartReader()
		if !assert.NoError(t, err) {
			return
		}

		file, err := reader.NextPart()
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "files", file.FormName())
		assert.Equal(t, "document.txt", file.FileName())

		contents, err := io.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, documentText, string(contents))
		assert.NoError(t, file.Close())

		strategy, err := reader.NextPart()
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, "strategy", strategy.FormName())

		value, err := io.ReadAll(strategy)
		assert.NoError(t, err)
		assert.Equal(t, "fast", string(value))
		assert.NoError(t, strategy.Close())
		w.Header().Set(contentTypeHeader, jsonContentType)
		_, err = io.WriteString(w, `[{
			"type":"NarrativeText","element_id":"element-1",
			"text":"A short document.",
			"metadata":{"filename":"document.txt","page_number":1}
		}]`)
		assert.NoError(t, err)
	})

	contentType, body := multipartDocument(t)
	res, err := client.PartitionWithBodyWithResponse(
		context.Background(), nil, contentType, body,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.Len(t, *res.JSON200, 1)

	element, err := json.Marshal((*res.JSON200)[0])
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"type":"NarrativeText","element_id":"element-1",
		"text":"A short document.",
		"metadata":{"filename":"document.txt","page_number":1}
	}`, string(element))
}

func TestPartitionValidationResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "array detail with mixed location values",
			body: `{"detail":[{
				"loc":["body","files",0],
				"msg":"File is required","type":"missing"
			}]}`,
		},
		{
			name: "string detail",
			body: `{"detail":"File is required"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t,
				func(w http.ResponseWriter, r *http.Request) {
					assertRequest(t, r)
					w.Header().Set(contentTypeHeader, jsonContentType)
					w.WriteHeader(http.StatusUnprocessableEntity)
					_, err := io.WriteString(w, tc.body)
					assert.NoError(t, err)
				},
			)

			contentType, body := multipartDocument(t)
			res, err := client.PartitionWithBodyWithResponse(
				context.Background(), nil, contentType, body,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode())
			require.NotNil(t, res.JSON422)

			detail, err := json.Marshal(res.JSON422)
			require.NoError(t, err)
			assert.JSONEq(t, tc.body, string(detail))
		})
	}
}

func TestPartitionServerError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r)
		w.Header().Set(contentTypeHeader, jsonContentType)
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"detail":"Processing failed"}`)
		assert.NoError(t, err)
	})

	contentType, body := multipartDocument(t)
	res, err := client.PartitionWithBodyWithResponse(
		context.Background(), nil, contentType, body,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.StatusCode())
	require.NotNil(t, res.JSON5XX)
	require.NotNil(t, res.JSON5XX.Detail)
	assert.Equal(t, "Processing failed", *res.JSON5XX.Detail)
}

func TestPartitionCSV(t *testing.T) {
	t.Parallel()

	const csvBody = "type,element_id,text\nNarrativeText,element-1,Text\n"

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r)
		w.Header().Set(contentTypeHeader, "text/csv")
		_, err := io.WriteString(w, csvBody)
		assert.NoError(t, err)
	})

	contentType, body := multipartDocument(t)
	res, err := client.PartitionWithBodyWithResponse(
		context.Background(), nil, contentType, body,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
	assert.Equal(t, csvBody, string(res.Body))
	assert.Nil(t, res.JSON200)
}

func TestPartitionMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r)
		w.Header().Set(contentTypeHeader, jsonContentType)
		_, err := io.WriteString(w, `[{"text":!}]`)
		assert.NoError(t, err)
	})

	_, err := client.PartitionWithBodyWithResponse(
		context.Background(), nil, jsonContentType, strings.NewReader("{}"),
	)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
}

func TestNullablePartitionOptions(t *testing.T) {
	t.Parallel()

	var options partition.PartitionParameters

	require.NoError(t, json.Unmarshal([]byte(`{
		"content_type":null,
		"starting_page_number":null,"include_orig_elements":null,
		"similarity_threshold":0.123456789012345
	}`), &options))
	assert.Nil(t, options.ContentType)
	assert.Nil(t, options.StartingPageNumber)
	assert.Nil(t, options.IncludeOrigElements)
	require.NotNil(t, options.SimilarityThreshold)
	assert.InDelta(t, 0.123456789012345, *options.SimilarityThreshold, 0)
}
