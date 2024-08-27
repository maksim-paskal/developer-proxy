package types

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
)

const ProxyRuleFormat = `^(prefix:|equal:|regexp:|)(/.+)@(https?://.+|endpoint)$`

type ProxyRuleOperator string

const (
	ProxyRuleOperatorPrefix ProxyRuleOperator = "prefix"
	ProxyRuleOperatorEqual  ProxyRuleOperator = "equal"
	ProxyRuleOperatorRegexp ProxyRuleOperator = "regexp"
)

type ProxyRule struct {
	Operator ProxyRuleOperator
	Value    string
	URL      string
}

func (r *ProxyRule) FromString(rule string) error {
	throwError := func(err error) {
		log.Fatal(err.Error() + ": parsing rule: " + rule + " (expected format: " + ProxyRuleFormat + ")")
	}

	re2 := regexp.MustCompile(ProxyRuleFormat)

	if !re2.MatchString(rule) {
		throwError(errors.New("invalid format"))
	}

	matches := re2.FindStringSubmatch(rule)

	r.Operator = ProxyRuleOperator(strings.TrimRight(matches[1], ":"))
	r.Value = matches[2]
	r.URL = matches[3]

	if r.Operator == "" {
		r.Operator = "prefix"
	}

	if err := r.Validate(); err != nil {
		throwError(err)
	}

	return nil
}

func (r *ProxyRule) Match(path string) bool {
	switch r.Operator {
	case ProxyRuleOperatorPrefix:
		return strings.HasPrefix(path, r.Value)
	case ProxyRuleOperatorEqual:
		return path == r.Value
	case ProxyRuleOperatorRegexp:
		return regexp.MustCompile(r.Value).MatchString(path)
	}

	return false
}

func (r *ProxyRule) Validate() error { //nolint:cyclop
	if r.Operator == "" {
		return errors.New("operator is required")
	}

	switch r.Operator {
	case ProxyRuleOperatorPrefix:
	case ProxyRuleOperatorEqual:
	case ProxyRuleOperatorRegexp:
	default:
		return errors.New("invalid operator")
	}

	if r.Operator == ProxyRuleOperatorRegexp {
		if _, err := regexp.Compile(r.Value); err != nil {
			return err
		}
	}

	if r.Value == "" {
		return errors.New("value is required")
	}

	if r.URL == "" {
		return errors.New("url is required")
	}

	if _, err := url.Parse(r.URL); err != nil {
		return err
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
