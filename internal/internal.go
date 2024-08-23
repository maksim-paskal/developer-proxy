package internal

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/maksim-paskal/developer-proxy/pkg/types"
)

func NewApplication() *Application {
	return &Application{
		HTTPClient: &http.Client{},
	}
}

type Application struct {
	Address    string
	Endpoint   string
	Rules      []types.ProxyRule
	Timeout    time.Duration
	HTTPClient *http.Client
}

func (a *Application) Validate() error {
	if a.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if _, err := url.Parse(a.Endpoint); err != nil {
		return err
	}

	return nil
}

func (a *Application) getTargetURL(currentPath string) string {
	for _, rule := range a.Rules {
		if !strings.HasPrefix(currentPath, rule.Prefix) {
			continue
		}

		if rule.URL == "endpoint" {
			return strings.TrimRight(a.Endpoint, "/") + currentPath
		}

		return strings.TrimRight(rule.URL, "/") + currentPath
	}

	return strings.TrimRight(a.Endpoint, "/") + currentPath
}

func (a *Application) handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), a.Timeout)
	defer cancel()

	targetURL := a.getTargetURL(r.URL.String())

	slog := slog.With(
		"method", r.Method,
		"url", r.URL.String(),
		"targetURL", targetURL,
	)

	slog.Debug("Proxying request")

	throwError := func(err error, msg string) {
		slog.Error(msg, "error", err.Error())
		http.Error(w, msg+": "+err.Error(), http.StatusInternalServerError)
	}

	// Create a new HTTP request with the same method, URL, and body as the original request
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		throwError(err, "Error creating proxy request")

		return
	}

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// Send the proxy request using the custom transport
	resp, err := a.HTTPClient.Do(proxyReq)
	if err != nil {
		throwError(err, "Error sending proxy request")

		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	slog.Debug("Proxy response received", "status", resp.StatusCode)

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		throwError(err, "Error copying response body")
	}
}

func (a *Application) Run() error {
	slog.Debug("Application starting", "application", a)

	server := http.Server{
		Addr:              a.Address,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           http.HandlerFunc(a.handleRequest),
	}

	slog.Info("Proxying requests to " + a.Endpoint)

	for _, rule := range a.Rules {
		slog.Info("Proxy rule: " + rule.Prefix + " -> " + rule.URL)
	}

	slog.Info("Proxy server listening on address http://" + server.Addr)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
