package rule

import (
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/relabel"
	"github.com/prometheus/prometheus/promql/parser"
)

type RelabelRuleConfig struct {
	Relabel []*relabel.Config `yaml:"relabel"`
}

type RelabelRule struct {
	Config RelabelRuleConfig
}

func convertMatchersToLabels(ms []*labels.Matcher) labels.Labels {
	ls := make(labels.Labels, 0)
	for _, m := range ms {
		ls = append(ls,
			labels.Label{
				Name:  m.Name,
				Value: m.Value,
			})
	}
	return ls
}

func mergeToMatcher(ls labels.Labels, ms []*labels.Matcher) []*labels.Matcher {
	nms := make([]*labels.Matcher, 0)
	for _, l := range ls {
		exist := false
		for _, m := range ms {
			if m.Name == l.Name {
				m.Value = l.Value
				nms = append(nms, m)
				exist = true
			}
		}
		if !exist {
			nms = append(nms, &labels.Matcher{
				Type:  labels.MatchEqual, // TODO: support regex matcher
				Name:  l.Name,
				Value: l.Value,
			})
		}
	}
	return nms
}

func (r *RelabelRule) Replace(expr parser.Expr) (parser.Expr, bool) {
	switch n := expr.(type) {
	case *parser.VectorSelector:
		ls := convertMatchersToLabels(n.LabelMatchers)
		pls := relabel.Process(ls, r.Config.Relabel...)
		n.LabelMatchers = mergeToMatcher(pls, n.LabelMatchers)
		for _, m := range n.LabelMatchers {
			if m.Name == "__name__" {
				n.Name = m.Value
			}
		}
	}
	return expr, false
}

func (r *RelabelRule) IsGenerateExpr() bool {
	return false
}
