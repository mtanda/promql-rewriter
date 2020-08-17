package rule

import (
	"regexp"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

type LabelJoinRuleConfig struct {
	NameMatcher    string   `yaml:"namematcher"`
	LabelProvider  string   `yaml:"label_provider"`
	TargetLabels   []string `yaml:"target_labels"`
	IgnoringLabels []string `yaml:"ignoring_labels"`
}

type LabelJoinRule struct {
	Config           LabelJoinRuleConfig
	NameMatcherRegex *regexp.Regexp
}

func (r *LabelJoinRule) Replace(expr parser.Expr) (parser.Expr, bool) {
	if r.NameMatcherRegex == nil {
		r.NameMatcherRegex = regexp.MustCompile(r.Config.NameMatcher)
	}
	switch n := expr.(type) {
	case *parser.VectorSelector:
		for _, m := range n.LabelMatchers {
			if m.Name == "__name__" && r.NameMatcherRegex.Match([]byte(m.Value)) {
				slm := make([]*labels.Matcher, 0)
				tlm := make([]*labels.Matcher, 0)
				for _, lm := range n.LabelMatchers {
					matched := false
					for _, tl := range r.Config.TargetLabels {
						if lm.Name == tl {
							tlm = append(tlm, lm)
							matched = true
						}
					}
					if !matched {
						slm = append(slm, lm)
					}
				}
				tlm = append(tlm, &labels.Matcher{
					Name:  "__name__",
					Value: r.Config.LabelProvider,
				})

				n.LabelMatchers = slm
				args := []parser.Expr{
					&parser.VectorSelector{
						Name:          r.Config.LabelProvider,
						LabelMatchers: tlm,
					},
					&parser.NumberLiteral{Val: 0},
				}
				fnc := &parser.Call{
					Func: parser.Functions["clamp_max"],
					Args: args,
				}
				return &parser.ParenExpr{
					Expr: &parser.BinaryExpr{
						Op:  parser.ADD,
						LHS: expr,
						RHS: fnc,
						VectorMatching: &parser.VectorMatching{
							Card:           parser.CardOneToMany,
							MatchingLabels: r.Config.IgnoringLabels,
							On:             false,
						},
						ReturnBool: false,
					},
				}, true
			}
		}
	}
	return expr, false
}

func (r *LabelJoinRule) IsGenerateExpr() bool {
	return true
}
