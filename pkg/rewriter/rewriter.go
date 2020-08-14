package rewriter

import (
	"reflect"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/mtanda/promql-rewriter/pkg/rule"
	"github.com/prometheus/prometheus/promql/parser"
)

type Rewriter struct {
	Rules  []rule.Rule
	Logger *log.Logger
}

func (r *Rewriter) replace(n *parser.VectorSelector, wrap bool, skip bool) (parser.Expr, bool) {
	var expr parser.Expr
	expr = n

	for _, rule := range r.Rules {
		if skip && rule.IsGenerateExpr() {
			continue
		}
		s := false
		expr, s = rule.Replace(expr)
		skip = skip || s
	}
	if wrap {
		expr = &parser.ParenExpr{Expr: expr}
	}
	return expr, skip
}

func (r *Rewriter) RewriteQuery(query string) (string, error) {
	expr, err := parser.ParseExpr(query)
	if err != nil {
		return "", err
	}
	skip := false
	if vs, ok := expr.(*parser.VectorSelector); ok {
		expr, skip = r.replace(vs, false, skip)
	}
	if _, ok := expr.(*parser.VectorSelector); !ok {
		f := func(node parser.Node, nodes []parser.Node, skip bool) (bool, error) {
			if node != nil {
				level.Debug(*r.Logger).Log("type", reflect.TypeOf(node), "string", node.String())
			}

			switch n := node.(type) {
			case *parser.BinaryExpr:
				if vs, ok := n.LHS.(*parser.VectorSelector); ok {
					n.LHS, skip = r.replace(vs, true, skip)
				}
				if vs, ok := n.RHS.(*parser.VectorSelector); ok {
					n.RHS, skip = r.replace(vs, true, skip)
				}
			case *parser.UnaryExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr, skip = r.replace(vs, true, skip)
				}
			case *parser.ParenExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr, skip = r.replace(vs, false, skip)
				}
			case *parser.AggregateExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr, skip = r.replace(vs, false, skip)
				}
			case *parser.SubqueryExpr:
				if vs, ok := n.Expr.(*parser.VectorSelector); ok {
					n.Expr, skip = r.replace(vs, false, skip)
				}
			case *parser.MatrixSelector:
				// not supported
			}
			return skip, nil
		}
		visitor := Inspector{process: f, skipGenerateExprRule: skip}
		parser.Walk(visitor, expr, nil)
	}

	replacedExpr := expr.String()
	_, err = parser.ParseExpr(replacedExpr)
	if err != nil {
		return "", err
	}
	return replacedExpr, nil
}

type Inspector struct {
	process              func(parser.Node, []parser.Node, bool) (bool, error)
	skipGenerateExprRule bool
}

func (f Inspector) Visit(node parser.Node, path []parser.Node) (parser.Visitor, error) {
	if skip, err := f.process(node, path, f.skipGenerateExprRule); err != nil {
		return nil, err
	} else {
		return Inspector{
			process:              f.process,
			skipGenerateExprRule: f.skipGenerateExprRule || skip,
		}, nil
	}
}
