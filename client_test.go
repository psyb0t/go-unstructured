package unstructured_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-unstructured"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey        = "EXAMPLE-DO-NOT-USE"
	apiKeyHeader      = "unstructured-api-key"
	contentTypeHeader = "Content-Type"
	jsonContentType   = "application/json"
	sourcePath        = "/api/v1/sources/"
	sourceName        = "Documents"
	s3ConnectorType   = "s3"
	sourceJSON        = `{
		"id":"00000000-0000-0000-0000-000000000001",
		"name":"Documents",
		"type":"s3",
		"config":{"remote_url":"s3://example-documents","anonymous":true},
		"created_at":"2026-10-17T00:00:00Z",
		"updated_at":null,
		"key":null
	}`
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *unstructured.ClientWithResponses {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	auth := unstructured.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(apiKeyHeader, testAPIKey)

			return nil
		},
	)
	client, err := unstructured.NewClientWithResponses(server.URL, auth)
	require.NoError(t, err)

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set(contentTypeHeader, jsonContentType)
	_, err := w.Write([]byte(body))
	assert.NoError(t, err)
}

func TestListSources(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, sourcePath, r.URL.Path)
		assert.Equal(t, s3ConnectorType, r.URL.Query().Get("source_type"))
		assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
		writeJSON(t, w, "["+sourceJSON+"]")
	})

	sourceType := unstructured.SourceConnectorType(s3ConnectorType)
	res, err := client.ListSourcesWithResponse(
		context.Background(),
		&unstructured.ListSourcesParams{SourceType: &sourceType},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.Len(t, *res.JSON200, 1)

	source := (*res.JSON200)[0]
	assert.Equal(t, sourceName, source.Name)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001",
		source.Id.String())
	assert.Nil(t, source.Key)
	assert.Nil(t, source.UpdatedAt)

	config, err := json.Marshal(source.Config)
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"remote_url":"s3://example-documents","anonymous":true}`,
		string(config),
	)
}

func TestCreateSource(t *testing.T) {
	t.Parallel()

	const requestJSON = `{
		"name":"Documents", "type":"s3",
		"config":{"remote_url":"s3://example-documents","anonymous":true}
	}`

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, sourcePath, r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
		assert.Equal(t, jsonContentType, r.Header.Get(contentTypeHeader))

		var body map[string]any

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, sourceName, body["name"])
		assert.Equal(t, s3ConnectorType, body["type"])
		assert.Equal(t, map[string]any{
			"remote_url": "s3://example-documents", "anonymous": true,
		}, body["config"])
		writeJSON(t, w, sourceJSON)
	})

	var body unstructured.CreateSourceJSONRequestBody

	require.NoError(t, json.Unmarshal([]byte(requestJSON), &body))

	res, err := client.CreateSourceWithResponse(
		context.Background(),
		&unstructured.CreateSourceParams{},
		body,
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Equal(t, sourceName, res.JSON200.Name)
}

func TestGetSourceEscapedID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, sourcePath+"source α", r.URL.Path)
		assert.Equal(t,
			sourcePath+"source%20%CE%B1",
			r.URL.EscapedPath(),
		)
		assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
		writeJSON(t, w, sourceJSON)
	})

	res, err := client.GetSourceWithResponse(
		context.Background(), "source α", nil,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	assert.Equal(t, sourceName, res.JSON200.Name)
}

func TestValidationResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, sourcePath, r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
		w.Header().Set(contentTypeHeader, jsonContentType)
		w.WriteHeader(http.StatusUnprocessableEntity)
		writeJSON(t, w, `{"detail":[{
			"loc":["query",0],"msg":"Invalid connector","type":"value_error"
		}]}`)
	})

	res, err := client.ListSourcesWithResponse(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode())
	require.NotNil(t, res.JSON422)
	require.NotNil(t, res.JSON422.Detail)
	require.Len(t, *res.JSON422.Detail, 1)

	detail := (*res.JSON422.Detail)[0]
	assert.Equal(t, "Invalid connector", detail.Msg)

	location, err := json.Marshal(detail.Loc)
	require.NoError(t, err)
	assert.JSONEq(t, `["query",0]`, string(location))
}

func TestMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `[{"name":!}]`)
	})

	_, err := client.ListSourcesWithResponse(context.Background(), nil)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
}

func TestUnauthenticatedResponse(t *testing.T) {
	t.Parallel()

	const responseBody = `{"detail":"Not authenticated"}`

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, sourcePath, r.URL.Path)
			assert.Empty(t, r.Header.Values(apiKeyHeader))
			w.Header().Set(contentTypeHeader, jsonContentType)
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(t, w, responseBody)
		},
	))
	t.Cleanup(server.Close)

	client, err := unstructured.NewClientWithResponses(server.URL)
	require.NoError(t, err)

	res, err := client.ListSourcesWithResponse(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode())
	assert.Nil(t, res.JSON200)
	assert.JSONEq(t, responseBody, string(res.Body))
}

func TestCanceledRequest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		assert.Fail(t, "Canceled request reached the server")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.ListSourcesWithResponse(ctx, nil)
	require.ErrorIs(t, err, context.Canceled)
}

func TestAdditionalProperties(t *testing.T) {
	t.Parallel()

	const payload = `{
		"api_endpoint":"https://example.invalid",
		"extension":{"nested":[null,true,9007199254740993]}
	}`

	const extension = `{"nested":[null,true,9007199254740993]}`

	var config unstructured.AstraDBConnectorConfig

	require.NoError(t, json.Unmarshal([]byte(payload), &config))
	assert.Equal(t, json.RawMessage(extension),
		config.AdditionalProperties["extension"])

	encoded, err := json.Marshal(config)
	require.NoError(t, err)

	var decoded map[string]json.RawMessage

	require.NoError(t, json.Unmarshal(encoded, &decoded))
	assert.Equal(t, json.RawMessage(extension), decoded["extension"])
}

func TestUnspecifiedJSONResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		body string
	}{
		{
			name: "object",
			body: `{"result":{"count":9007199254740993}}`,
		},
		{name: "array", body: `[true,null,"accepted"]`},
		{name: "scalar", body: `"accepted"`},
		{name: "null", body: `null`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t,
				func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodDelete, r.Method)
					assert.Equal(t, sourcePath+"source-1", r.URL.Path)
					assert.Equal(t, testAPIKey, r.Header.Get(apiKeyHeader))
					writeJSON(t, w, tc.body)
				},
			)
			res, err := client.DeleteSourceWithResponse(
				context.Background(), "source-1", nil,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode())
			require.NotNil(t, res.JSON200)
			assert.Equal(t, json.RawMessage(tc.body), *res.JSON200)
		})
	}
}

func TestRequestEditorError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		assert.Fail(t, "Rejected request reached the server")
	})

	_, err := client.ListSourcesWithResponse(
		context.Background(),
		nil,
		func(_ context.Context, _ *http.Request) error {
			return context.Canceled
		},
	)
	require.ErrorIs(t, err, context.Canceled)
}
