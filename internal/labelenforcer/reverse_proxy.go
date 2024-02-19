package labelenforcer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(backendURL *url.URL, labels []string) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.Transport = NewEnforcer(labels, nil)

	return proxy
}
