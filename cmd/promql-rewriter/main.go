package main

import (
	"flag"
	"io/ioutil"

	"github.com/go-kit/kit/log/level"
	"github.com/mtanda/promql-rewriter/pkg/rewriter"
	"github.com/mtanda/promql-rewriter/pkg/rule"
	"github.com/prometheus/common/promlog"
	"gopkg.in/yaml.v2"
)

func loadRuleConfigs(configFile string) ([]rule.RuleConfig, error) {
	var ruleConfigs []rule.RuleConfig

	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(buf, &ruleConfigs)
	if err != nil {
		return nil, err
	}

	return ruleConfigs, nil
}

func loadRules(ruleConfigs []rule.RuleConfig) []rule.Rule {
	rules := make([]rule.Rule, 0)
	for _, cfg := range ruleConfigs {
		if cfg.CounterConfig != nil {
			rules = append(rules, &rule.CounterRule{Config: *cfg.CounterConfig})
		}
		if cfg.NameMapConfig != nil {
			rules = append(rules, &rule.NameMapRule{Config: *cfg.NameMapConfig})
		}
		if cfg.RelabelConfig != nil {
			rules = append(rules, &rule.RelabelRule{Config: *cfg.RelabelConfig})
		}
	}
	return rules
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config.file", "./promql-rewriter.yaml", "Configuration file path.")
	flag.Parse()

	logLevel := promlog.AllowedLevel{}
	logLevel.Set("info")
	format := promlog.AllowedFormat{}
	format.Set("text")
	logCfg := promlog.Config{Level: &logLevel, Format: &format}
	logger := promlog.New(&logCfg)

	query := flag.Arg(0)
	level.Info(logger).Log("original query", query)

	ruleConfigs, err := loadRuleConfigs(configFile)
	if err != nil {
		panic(err)
	}
	rules := loadRules(ruleConfigs)
	rewriter := &rewriter.Rewriter{Rules: rules, Logger: &logger}
	rewritedQuery, err := rewriter.RewriteQuery(query)
	if err != nil {
		panic(err)
	}
	level.Info(logger).Log("rewrited query", rewritedQuery)
}
