package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/maksim-paskal/developer-proxy/internal"
	"github.com/maksim-paskal/developer-proxy/pkg/types"
)

var (
	version  = flag.Bool("version", false, "Print the version and exit")
	debug    = flag.Bool("debug", false, "Enable debug logging")
	address  = flag.String("address", "127.0.0.1:10000", "Proxy server address")
	timeout  = flag.Duration("timeout", time.Minute, "The timeout for proxy requests")
	endpoint = flag.String("endpoint", "", "The endpoint to proxy requests to")
	rules    types.ProxyRules
)

func main() {
	flag.Var(&rules, "rule", "Rule to route proxy requests (format: "+types.ProxyRuleFormat+") (can be specified multiple times)")
	flag.Parse()

	if *version {
		fmt.Println(internal.Version)
		return
	}

	if *debug {
		slog.Info("Debug logging enabled")
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	application := internal.NewApplication()
	application.Address = *address
	application.Timeout = *timeout
	application.Endpoint = *endpoint
	application.Rules = rules

	if err := application.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
