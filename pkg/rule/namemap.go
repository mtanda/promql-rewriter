package rule

import (
	"github.com/prometheus/prometheus/promql/parser"
)

type NameMapRuleConfig struct {
	NameMap map[string]string `yaml:"namemap"`
}

type NameMapRule struct {
	Config NameMapRuleConfig
}

func (r *NameMapRule) Replace(expr parser.Expr) parser.Expr {
	switch n := expr.(type) {
	case *parser.VectorSelector:
		for i, m := range n.LabelMatchers {
			if m.Name == "__name__" {
				origName := n.LabelMatchers[i].Value
				if v, ok := r.Config.NameMap[origName]; ok {
					n.LabelMatchers[i].Value = v
					n.Name = v
				}
			}
		}
	}
	return expr
}
