package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProxyMode string

var (
	debug      = flag.Bool("debug", false, "Enable debug logging")
	address    = flag.String("address", "127.0.0.1:10000", "Proxy server address")
	timeout    = flag.Duration("timeout", 60*time.Second, "The timeout for proxy requests")
	endpoint   = flag.String("endpoint", "", "The endpoint to proxy requests to")
	proxyRules = flag.String("rules", "", "Comma separated list of rules to proxy requests")
)

func NewApplication() *Application {
	return &Application{
		Address:    *address,
		Endpoint:   *endpoint,
		ProxyRules: *proxyRules,
		Timeout:    *timeout,
		HTTPClient: &http.Client{},
	}
}

type Application struct {
	Address    string
	Endpoint   string
	ProxyRules string
	rules      []ProxyRule
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
	for _, rule := range a.rules {
		if !strings.HasPrefix(currentPath, rule.Prefix) {
			continue
		}

		if rule.URL == "endpoint" {
			return strings.TrimRight(*endpoint, "/") + currentPath
		}

		return strings.TrimRight(rule.URL, "/") + currentPath
	}

	return strings.TrimRight(*endpoint, "/") + currentPath
}

func (a *Application) handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	targetURL := a.getTargetURL(r.URL.String())

	slog.Debug("Proxying request to " + targetURL)

	ctx, cancel := context.WithTimeout(ctx, a.Timeout)
	defer cancel()

	// Create a new HTTP request with the same method, URL, and body as the original request
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)

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
		slog.Error(err.Error())
		http.Error(w, "Error sending proxy request: "+targetURL, http.StatusInternalServerError)

		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "Error copying response body", http.StatusInternalServerError)
	}
}

type ProxyRule struct {
	Prefix string
	URL    string
}

func (r *ProxyRule) Validate() error {
	if r.Prefix == "" {
		return errors.New("prefix is required")
	}

	if _, err := url.Parse(r.URL); err != nil {
		return err
	}

	validURL := func() bool {
		if strings.HasPrefix(r.URL, "http://") {
			return true
		}

		if strings.HasPrefix(r.URL, "https://") {
			return true
		}

		if r.URL == "endpoint" {
			return true
		}

		return false
	}

	if !validURL() {
		return errors.New("url must start with http:// or https:// or endpoint")
	}

	return nil
}

func (a *Application) parseProxyRules(rules string) ([]ProxyRule, error) {
	result := make([]ProxyRule, 0)

	if rules == "" {
		return result, nil
	}

	makeErrror := func(rule string, err error) error {
		return errors.New(err.Error() + ": invalid proxy rule: " + rule + " (expected format: /path@http://target)")
	}

	for _, rule := range strings.Split(rules, ",") {
		slog.Debug("Parsing proxy rule: " + rule)

		parts := strings.Split(rule, "@")
		if partsLen := len(parts); partsLen != 2 {
			return nil, makeErrror(rule, errors.New("invalid parts"))
		}

		proxyRule := ProxyRule{
			Prefix: parts[0],
			URL:    parts[1],
		}

		if err := proxyRule.Validate(); err != nil {
			return nil, makeErrror(rule, err)
		}

		result = append(result, proxyRule)
	}

	return result, nil
}

func (a *Application) Run() error {
	server := http.Server{
		Addr:              *address,
		ReadHeaderTimeout: 10 * time.Second,
	}

	http.Handle("/", http.HandlerFunc(a.handleRequest))

	slog.Info("Proxying requests to " + *endpoint)

	rules, err := a.parseProxyRules(*proxyRules)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		slog.Info("Proxy rule: " + rule.Prefix + " -> " + rule.URL)
	}

	a.rules = rules

	slog.Info("Proxy server listening on address http://" + server.Addr)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	if *debug {
		slog.Info("Debug logging enabled")
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	application := NewApplication()

	if err := application.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
