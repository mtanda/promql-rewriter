package rule

import (
	"github.com/prometheus/prometheus/promql/parser"
)

type RuleConfig struct {
	CounterConfig   *CounterRuleConfig   `yaml:"counter_rule_config,omitempty"`
	LabelJoinConfig *LabelJoinRuleConfig `yaml:"labeljoin_rule_config,omitempty"`
	NameMapConfig   *NameMapRuleConfig   `yaml:"namemap_rule_config,omitempty"`
	RelabelConfig   *RelabelRuleConfig   `yaml:"relabel_rule_config,omitempty"`
}

type Rule interface {
	Replace(expr parser.Expr) (parser.Expr, bool)
	IsGenerateExpr() bool
}
