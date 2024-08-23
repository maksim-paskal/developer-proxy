package types

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
)

const ProxyRuleFormat = "/path@http://target"

type ProxyRule struct {
	Prefix string
	URL    string
}

func (r *ProxyRule) FromString(rule string) error {
	throwError := func(err error) {
		log.Fatal(err.Error() + ": parsing rule: " + rule + " (expected format: " + ProxyRuleFormat + ")")
	}

	parts := strings.Split(rule, "@")
	if len(parts) != 2 {
		throwError(errors.New("invalid parts"))
	}

	r.Prefix = parts[0]
	r.URL = parts[1]

	if err := r.Validate(); err != nil {
		throwError(err)
	}

	return nil
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

type ProxyRules []ProxyRule

func (p *ProxyRules) String() string {
	return fmt.Sprintf("%v", *p)
}

func (p *ProxyRules) Set(value string) error {
	rule := ProxyRule{}
	if err := rule.FromString(value); err != nil {
		return err
	}

	*p = append(*p, rule)

	return nil
}
