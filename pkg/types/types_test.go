package types_test

import (
	"testing"

	"github.com/maksim-paskal/developer-proxy/pkg/types"
)

func TestRule(t *testing.T) { //nolint:funlen
	t.Parallel()

	type test struct {
		Name  string
		Rule  types.ProxyRule
		Path  string
		Match bool
	}

	tests := []test{
		{
			Name: "Match prefix",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorPrefix,
				Value:    "/api",
			},
			Path:  "/api/v1",
			Match: true,
		},
		{
			Name: "Match prefix (negative)",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorPrefix,
				Value:    "/api",
			},
			Path:  "/test",
			Match: false,
		},
		{
			Name: "Match equal",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorEqual,
				Value:    "/api",
			},
			Path:  "/api",
			Match: true,
		},
		{
			Name: "Match equal (negative)",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorEqual,
				Value:    "/api",
			},
			Path:  "/api/test",
			Match: false,
		},
		{
			Name: "Match regexp",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorRegexp,
				Value:    "^/(api|test|v1)",
			},
			Path:  "/api",
			Match: true,
		},
		{
			Name: "Match regexp (negative)",
			Rule: types.ProxyRule{
				Operator: types.ProxyRuleOperatorRegexp,
				Value:    "^/(api|test|v1)",
			},
			Path:  "/ap3/test",
			Match: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			if match := test.Rule.Match(test.Path); match != test.Match {
				t.Errorf("Expected %v, got %v", test.Match, match)
			}
		})
	}
}

func TestBadRule(t *testing.T) {
	t.Parallel()

	test := []types.ProxyRule{
		{
			Operator: "bad",
			Value:    "/api",
		},
		{
			Operator: "regexp",
			Value:    "?!re",
		},
	}

	for _, rule := range test {
		err := rule.Validate()
		t.Log(err)

		if err == nil {
			t.Errorf("Expected false, got true")
		}
	}
}
