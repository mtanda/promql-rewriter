package rewriter

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/mtanda/promql-rewriter/pkg/rule"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/prometheus/promql/parser"
)

func TestRewriter_RewriteQuery(t *testing.T) {
	logLevel := promlog.AllowedLevel{}
	logLevel.Set("error")
	format := promlog.AllowedFormat{}
	format.Set("text")
	logCfg := promlog.Config{Level: &logLevel, Format: &format}
	logger := promlog.New(&logCfg)
	type fields struct {
		Rules         []rule.Rule
		GeneratedExpr []parser.Expr
		Logger        *log.Logger
	}
	type args struct {
		query string
	}
	f := fields{
		Rules: []rule.Rule{
			&rule.NameMapRule{
				Config: rule.NameMapRuleConfig{
					NameMap: map[string]string{
						"foo": "bar",
					},
				},
			},
			&rule.LabelJoinRule{
				Config: rule.LabelJoinRuleConfig{
					NameMatcher:    "label_join_target",
					LabelProvider:  "label_provider",
					TargetLabels:   []string{"tl1", "tl2"},
					IgnoringLabels: []string{"ig1", "ig2"},
				},
			},
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "single vector selector",
			fields:  f,
			args:    args{query: "foo"},
			want:    "bar",
			wantErr: false,
		},
		{
			name:    "single matrix selector",
			fields:  f,
			args:    args{query: "foo[1m]"},
			want:    "foo[1m]", // not supported, don't rewrite query
			wantErr: false,
		},
		{
			name:    "binary expression (lhs)",
			fields:  f,
			args:    args{query: "foo + 1"},
			want:    "(bar) + 1",
			wantErr: false,
		},
		{
			name:    "binary expression (rhs)",
			fields:  f,
			args:    args{query: "1 + foo"},
			want:    "1 + (bar)",
			wantErr: false,
		},
		{
			name:    "unary expression",
			fields:  f,
			args:    args{query: "-foo"},
			want:    "-(bar)",
			wantErr: false,
		},
		{
			name:    "paren expression",
			fields:  f,
			args:    args{query: "(foo)"},
			want:    "(bar)",
			wantErr: false,
		},
		{
			name:    "aggregate expression",
			fields:  f,
			args:    args{query: "avg (foo) by (bar)"},
			want:    "avg by(bar) (bar)",
			wantErr: false,
		},
		{
			name:    "subquery expression",
			fields:  f,
			args:    args{query: "avg_over_time(foo[30m:1m])"},
			want:    "avg_over_time(bar[30m:1m])",
			wantErr: false,
		},
		{
			name:    "label join",
			fields:  f,
			args:    args{query: "label_join_target{tl1=\"vtl1\",tl2=\"vtl2\",ig1=\"vig1\",ig2=\"vig2\"}"},
			want:    "label_join_target{ig1=\"vig1\",ig2=\"vig2\"} + ignoring(ig1, ig2) group_right() clamp_max(label_provider{tl1=\"vtl1\",tl2=\"vtl2\"}, 0)",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rewriter{
				Rules:         tt.fields.Rules,
				GeneratedExpr: tt.fields.GeneratedExpr,
				Logger:        &logger,
			}
			got, err := r.RewriteQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Rewriter.RewriteQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Rewriter.RewriteQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
