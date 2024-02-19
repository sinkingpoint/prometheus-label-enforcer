package labelenforcer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// PrometheusError is the error response format for Prometheus.
// See: https://prometheus.io/docs/prometheus/latest/querying/api/
type PrometheusError struct {
	Status    string `json:"status"`
	ErrorType string `json:"errorType"`
	Error     string `json:"error"`
}

func NewPrometheusError(err error) *http.Response {
	out, _ := json.Marshal(PrometheusError{
		Status:    "error",
		ErrorType: "not_acceptable",
		Error:     err.Error(),
	})

	return &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewReader(out)),
	}
}
