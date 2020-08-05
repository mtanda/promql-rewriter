package rewriter

import (
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mtanda/promql-rewriter/pkg/rule"
	"github.com/prometheus/prometheus/promql/parser"
)

type Rewriter struct {
	Rules         []rule.Rule
	GeneratedExpr []parser.Expr
	Logger        *log.Logger
}

func (r *Rewriter) replace(n *parser.VectorSelector, wrap bool) parser.Expr {
	var e parser.Expr
	e = n
	for _, rule := range r.Rules {
		e = rule.Replace(e)
	}
	if wrap {
		e = &parser.ParenExpr{Expr: e}
	}
	r.GeneratedExpr = append(r.GeneratedExpr, e)
	return e
}

func (r *Rewriter) RewriteQuery(query string) (string, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return "", err
	}
	if vs, ok := expr.(*parser.VectorSelector); !ok {
		parser.Inspect(expr, func(node parser.Node, nodes []parser.Node) error {
			for _, ge := range r.GeneratedExpr {
				if node == ge {
					return nil
				}
			}

			if node != nil {
				level.Debug(*r.Logger).Log("type", reflect.TypeOf(node), "string", node.String())
			}

			switch n := node.(type) {
			case *parser.BinaryExpr:
				if vs, ok := n.LHS.(*parser.VectorSelector); ok {
					n.LHS = r.replace(vs, true)
				}
				if vs, ok := n.RHS.(*parser.VectorSelector); ok {
					n.RHS = r.replace(vs, true)
				}
			case *parser.UnaryExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr = r.replace(vs, true)
				}
			case *parser.ParenExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr = r.replace(vs, false)
				}
			case *parser.AggregateExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr = r.replace(vs, false)
				}
			case *parser.SubqueryExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr = r.replace(vs, false)
				}
			case *parser.MatrixSelector:
				// not supported
			}
			return nil
		})
	} else {
		expr = r.replace(vs, false)
	}

	replacedExpr := expr.String()
	_, err = parser.ParseExpr(replacedExpr)
	if err != nil {
		return "", err
	}
	return replacedExpr, nil
}
