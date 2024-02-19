package main

import (
	"net/http"
	"net/url"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
	labelenforcer "github.com/sinkingpoint/label-enforcer/internal/labelenforcer"
)

var CLI struct {
	ListenAddress string   `default:":4278" help:"Address to listen on for HTTP requests."`
	BackendURL    string   `default:"http://:9090" help:"URL of the backend to proxy requests to."`
	Labels        []string `required:"true" help:"Comma-separated list of labels to enforce."`
}

func main() {
	kong.ConfigureHelp(kong.HelpOptions{
		Compact: true,
	})

	kong.Parse(&CLI)

	backendURL, err := url.Parse(CLI.BackendURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse backend url")
	}

	proxy := labelenforcer.NewReverseProxy(backendURL, CLI.Labels)

	server := http.Server{
		Addr:    CLI.ListenAddress,
		Handler: proxy,
	}

	log.Info().Str("address", CLI.ListenAddress).Str("backend", CLI.BackendURL).Strs("labels", CLI.Labels).Msg("starting label enforcer")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server failed")
	}
}
