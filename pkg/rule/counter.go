package rule

import (
	"regexp"
	"time"

	"github.com/prometheus/prometheus/promql/parser"
)

type CounterRuleConfig struct {
	NameMatcher string `yaml:"namematcher"`
	Range       string `yaml:"range"`
}

type CounterRule struct {
	Config           CounterRuleConfig
	NameMatcherRegex *regexp.Regexp
}

func (r *CounterRule) Replace(expr parser.Expr) parser.Expr {
	if r.NameMatcherRegex == nil {
		r.NameMatcherRegex = regexp.MustCompile(r.Config.NameMatcher)
	}
	switch n := expr.(type) {
	case *parser.VectorSelector:
		rng, err := time.ParseDuration(r.Config.Range)
		if err != nil {
			panic(err)
		}
		for _, m := range n.LabelMatchers {
			if m.Name == "__name__" && r.NameMatcherRegex.Match([]byte(m.Value)) {
				args := []parser.Expr{
					&parser.MatrixSelector{
						VectorSelector: n,
						Range:          rng,
					},
				}
				return &parser.Call{
					Func: parser.Functions["rate"],
					Args: args,
				}
			}
		}
	}
	return expr
}
