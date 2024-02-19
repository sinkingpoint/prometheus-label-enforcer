package labelenforcer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy creates a new reverse proxy that enforces the given labels.
func NewReverseProxy(backendURL *url.URL, labels []string) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.Transport = NewEnforcer(labels, nil)

	return proxy
}
