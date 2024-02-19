package labelenforcer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/VictoriaMetrics/metricsql"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

const queryParamName = "query"
const queryURLPath = "/api/v1/query"

var _ http.RoundTripper = &Enforcer{}

type Enforcer struct {
	labels           []string
	labelSet         map[string]struct{}
	backingTransport *http.Transport
}

func NewEnforcer(labels []string, backingTransport *http.Transport) *Enforcer {
	if backingTransport == nil {
		backingTransport = http.DefaultTransport.(*http.Transport)
	}

	labelSet := make(map[string]struct{}, len(labels))
	for _, label := range labels {
		labelSet[label] = struct{}{}
	}

	return &Enforcer{
		labels:           labels,
		labelSet:         labelSet,
		backingTransport: backingTransport,
	}
}

func (e *Enforcer) getQuery(req *http.Request) (string, error) {
	var body []byte
	var err error
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		defer func() {
			req.Body.Close()
			req.Body = io.NopCloser(bytes.NewReader(body))
		}()

		if err != nil {
			return "", err
		}
	}

	if req.Method == http.MethodGet {
		return req.URL.Query().Get(queryParamName), nil
	} else if req.Method == http.MethodPost {
		params, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}

		return params.Get(queryParamName), nil
	} else {
		return "", fmt.Errorf("unsupported method: %s", req.Method)
	}
}

func (e *Enforcer) HasLabels(query string) error {
	expr, err := metricsql.Parse(query)
	if err != nil {
		return fmt.Errorf("failed to parse query expression: %w", err)
	}

	var missingErr error
	metricsql.VisitAll(expr, func(expr metricsql.Expr) {
		if m, ok := expr.(*metricsql.MetricExpr); ok {
			for _, filterss := range m.LabelFilterss {
				for _, filter := range filterss {
					if _, ok := e.labelSet[filter.Label]; ok {
						return
					}
				}
			}

			// We didn't find any label filters for this metric.
			exprString := m.AppendString(nil)
			err := fmt.Errorf("%s is missing a label filter, expected at least one of: %s", exprString, strings.Join(e.labels, ", "))
			missingErr = multierr.Append(missingErr, err)
		}
	})

	return missingErr
}

func (e *Enforcer) enforce(req *http.Request) error {
	if req.URL.Path != queryURLPath {
		return nil
	}

	query, err := e.getQuery(req)
	if err != nil {
		return fmt.Errorf("failed to extract query from request: %w", err)
	}

	err = e.HasLabels(query)
	if err != nil {
		log.Debug().Err(err).Msgf("rejecting %q due to missing label filters", query)
	}

	return err
}

func (e *Enforcer) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := e.enforce(req); err != nil {
		return NewPrometheusError(err), nil
	}

	return e.backingTransport.RoundTrip(req)
}
