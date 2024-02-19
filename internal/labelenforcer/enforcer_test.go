package labelenforcer_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sinkingpoint/label-enforcer/internal/labelenforcer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelEnforcerLabels(t *testing.T) {
	enforcer := labelenforcer.NewEnforcer([]string{"foo", "test"}, nil)
	require.NotNil(t, enforcer)

	assert.NoError(t, enforcer.HasLabels("my_metric{foo=\"bar\"}"))
	assert.NoError(t, enforcer.HasLabels("my_metric{test=\"bar\"}"))
	assert.NoError(t, enforcer.HasLabels("my_metric{foo=\"bar\",bar=\"foo\"}"))
	assert.NoError(t, enforcer.HasLabels("my_metric{bar=\"foo\",foo=\"bar\"} + 1"))
	assert.NoError(t, enforcer.HasLabels("my_metric{bar=\"foo\",foo=\"bar\"} + my_metric{foo=\"bar\"}"))
	assert.Error(t, enforcer.HasLabels("my_metric{bar=\"foo\"}"))
	assert.Error(t, enforcer.HasLabels("my_metric"))
	assert.Error(t, enforcer.HasLabels("my_metric{}"))
	assert.Error(t, enforcer.HasLabels("my_metric{foo=\"bar\",bar=\"foo\"} + my_metric{bar=\"foo\"}"))
}

func newGetRequest(t *testing.T, query string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "/api/v1/query", nil)
	if err != nil {
		require.FailNow(t, "failed to create request", err)
	}

	queryBody := req.URL.Query()
	queryBody.Set("query", query)
	req.URL.RawQuery = queryBody.Encode()

	return req
}

func newPostRequest(t *testing.T, query string) *http.Request {
	queryBody := url.Values{}
	queryBody.Set("query", query)
	body := queryBody.Encode()

	req, err := http.NewRequest(http.MethodPost, "/api/v1/query", strings.NewReader(body))
	if err != nil {
		require.FailNow(t, "failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

func TestLabelEnforcerHTTP(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		query         string
		expectSuccess bool
	}{
		{
			name:          "get with valid query",
			method:        http.MethodGet,
			query:         "my_metric{foo=\"bar\"}",
			expectSuccess: true,
		},
		{
			name:          "get with invalid query",
			method:        http.MethodGet,
			query:         "my_metric{bar=\"foo\"}",
			expectSuccess: false,
		},
		{
			name:          "post with valid query",
			method:        http.MethodPost,
			query:         "my_metric{foo=\"bar\"}",
			expectSuccess: true,
		},
		{
			name:          "post with invalid query",
			method:        http.MethodPost,
			query:         "my_metric{bar=\"foo\"}",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectSuccess {
					require.Equal(t, tt.query, r.FormValue("query"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			backendURL, err := url.Parse(backend.URL)
			require.NoError(t, err)

			enforcer := labelenforcer.NewReverseProxy(backendURL, []string{"foo", "test"})
			require.NotNil(t, enforcer)

			var req *http.Request
			if tt.method == http.MethodGet {
				req = newGetRequest(t, tt.query)
			} else if tt.method == http.MethodPost {
				req = newPostRequest(t, tt.query)
			} else {
				require.FailNow(t, "unsupported method", tt.method)
			}

			writer := httptest.NewRecorder()
			enforcer.ServeHTTP(writer, req)

			if tt.expectSuccess {
				assert.Equal(t, http.StatusOK, writer.Code)
				assert.Equal(t, "OK", writer.Body.String())
			} else {
				assert.Equal(t, http.StatusBadRequest, writer.Code)
			}
		})
	}
}
