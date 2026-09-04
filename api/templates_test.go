package api

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"github.com/block/iterable-go/errors"
	"github.com/block/iterable-go/logger"
	"github.com/block/iterable-go/rate"
	"github.com/block/iterable-go/types"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplatesApi(t *testing.T) {
	client := &http.Client{}
	api := NewTemplatesApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	assert.NotNil(t, api)
	assert.NotNil(t, api.api)
	assert.Equal(t, testApiKey, api.api.apiKey)
	assert.Equal(t, client, api.api.httpClient)
}

func TestTemplates_Get(t *testing.T) {
	testCases := []struct {
		name       string
		resBody    []byte
		resCode    int
		resErr     error
		expectRes  *types.TemplatesResponse
		expectErr  bool
		resErrType string
	}{
		{
			name: "successful page",
			resBody: []byte(`{
				"templates": [{
					"templateId": 1,
					"name": "Welcome",
					"createdAt": "2026-09-01 12:00:00 +00:00",
					"updatedAt": "2026-09-02 12:00:00 +00:00",
					"creatorUserId": "creator@example.com",
					"messageTypeId": 2
				}],
				"nextPageUrl": "/api/templates?page=2&pageSize=1000",
				"totalTemplatesCount": 1
			}`),
			resCode: http.StatusOK,
			expectRes: &types.TemplatesResponse{
				Templates: []types.Template{{
					TemplateId: 1, Name: "Welcome", CreatedAt: "2026-09-01 12:00:00 +00:00",
					UpdatedAt: "2026-09-02 12:00:00 +00:00", CreatorUserId: "creator@example.com", MessageTypeId: 2,
				}},
				NextPageUrl:         "/api/templates?page=2&pageSize=1000",
				TotalTemplatesCount: 1,
			},
		},
		{name: "malformed JSON", resBody: []byte(`{"templates":[{]}`), resCode: http.StatusOK, expectErr: true, resErrType: errors.TYPE_JSON_PARSE},
		{name: "server error", resBody: []byte(`{"message":"error"}`), resCode: http.StatusInternalServerError, expectErr: true, resErrType: errors.TYPE_HTTP_STATUS},
		{name: "network error", resErr: assert.AnError, expectErr: true, resErrType: errors.TYPE_IO},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := httpClient(tt.resBody, tt.resCode, tt.resErr)
			api := NewTemplatesApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

			response, err := api.Get(1, 1000)
			if tt.expectErr {
				assert.Error(t, err)
				apiError := err.(*errors.ApiError)
				assert.Equal(t, tt.resErrType, apiError.Type)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectRes, response)
			}

			transport := client.Transport.(*testTransport)
			assert.Equal(t, "https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id", transport.Url())
			assert.Equal(t, http.MethodGet, transport.Method())
			assert.Equal(t, testApiKey, transport.ApiKey())
		})
	}
}

func TestTemplates_AllPaginates(t *testing.T) {
	t.Parallel()

	transport := &templatePagesTransport{bodies: [][]byte{
		[]byte(`{"templates":[{"templateId":1,"name":"First"}],"nextPageUrl":"/api/templates?page=2&pageSize=1000"}`),
		[]byte(`{"templates":[{"templateId":2,"name":"Second"}]}`),
	}}
	client := &http.Client{Transport: transport}
	api := NewTemplatesApi(testApiKey, client, &logger.Noop{}, &rate.NoopLimiter{})

	templates, err := api.All()
	assert.NoError(t, err)
	assert.Equal(t, []types.Template{{TemplateId: 1, Name: "First"}, {TemplateId: 2, Name: "Second"}}, templates)
	assert.Equal(t, []string{
		"https://api.iterable.com/api/templates?page=1&pageSize=1000&sort=id",
		"https://api.iterable.com/api/templates?page=2&pageSize=1000&sort=id",
	}, transport.urls)
}

func TestTemplates_GetRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	api := NewTemplatesApi(testApiKey, &http.Client{}, &logger.Noop{}, &rate.NoopLimiter{})
	testCases := []struct {
		name     string
		page     int
		pageSize int
	}{
		{name: "page below one", page: 0, pageSize: 1000},
		{name: "page size below one", page: 1, pageSize: 0},
		{name: "page size above limit", page: 1, pageSize: 1001},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := api.Get(tt.page, tt.pageSize)
			assert.Error(t, err)
		})
	}
}

type templatePagesTransport struct {
	mu     sync.Mutex
	bodies [][]byte
	urls   []string
}

func (t *templatePagesTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.urls = append(t.urls, request.URL.String())
	body := t.bodies[0]
	t.bodies = t.bodies[1:]
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &testReader{Reader: bytes.NewReader(body)},
	}, nil
}
