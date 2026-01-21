package linters

import (
	"log"
	"os"
	"sort"

	"github.com/googleapis/api-linter/v2/lint"
	"github.com/googleapis/api-linter/v2/rules"
)

var globalRules = lint.NewRuleRegistry()

func init() {
	if err := rules.Add(globalRules); err != nil {
		log.Fatalf("error when registering rules: %v", err)
	}
}

type (
	listedRule struct {
		Name lint.RuleName
	}
	listedRules       []listedRule
	listedRulesByName []listedRule
)

func (a listedRulesByName) Len() int           { return len(a) }
func (a listedRulesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a listedRulesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func outputRules(formatType string) error {
	rules := listedRules{}
	for id := range globalRules {
		rules = append(rules, listedRule{
			Name: id,
		})
	}

	sort.Sort(listedRulesByName(rules))

	// Determine the format for printing the results.
	// YAML format is the default.
	marshal := getOutputFormatFunc(formatType)

	// Print the results.
	b, err := marshal(rules)
	if err != nil {
		return err
	}
	w := os.Stdout
	if _, err = w.Write(b); err != nil {
		return err
	}

	return nil
}
