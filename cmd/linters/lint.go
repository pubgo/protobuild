// Package linters provides proto file linting with AIP rules.
package linters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/googleapis/api-linter/v2/lint"
	"github.com/pubgo/redant"
	"github.com/samber/lo"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gopkg.in/yaml.v3"

	"github.com/pubgo/protobuild/internal/typex"
)

// CliArgs holds command line arguments for the linter.
type CliArgs struct {
	// FormatType string
	// ProtoImportPaths          []string
	EnabledRules  []string
	DisabledRules []string
	ListRulesFlag bool
	DebugFlag     bool
	// IgnoreCommentDisablesFlag bool
}

// NewCli creates a new CLI arguments instance and options.
func NewCli() (*CliArgs, typex.Options) {
	var cliArgs CliArgs

	return &cliArgs, typex.Options{
		//redant.Option{
		//	Flag:        "ignore-comment-disables",
		//	Description: "If set to true, disable comments will be ignored.\nThis is helpful when strict enforcement of AIPs are necessary and\nproto definitions should not be able to disable checks.",
		//	Value:       redant.BoolOf(&cliArgs.IgnoreCommentDisablesFlag),
		//},

		redant.Option{
			Flag:        "debug",
			Description: "Run in debug mode. Panics will print stack.",
			Value:       redant.BoolOf(&cliArgs.DebugFlag),
		},

		redant.Option{
			Flag:        "list-rules",
			Description: "Print the rules and exit.  Honors the output-format flag.",
			Value:       redant.BoolOf(&cliArgs.ListRulesFlag),
		},

		//&cli.StringFlag{
		//	Name:        "output-format",
		//	Usage:       "The format of the linting results.\nSupported formats include \"yaml\", \"json\",\"github\" and \"summary\" table.\nYAML is the default.",
		//	Aliases:     []string{"f"},
		//	Value:       "",
		//	Destination: &cliArgs.FormatType,
		//},

		//&cli.StringSliceFlag{
		//	Name:        "proto-path",
		//	Usage:       "The folder for searching proto imports.\\nMay be specified multiple times; directories will be searched in order.\\nThe current working directory is always used.",
		//	Aliases:     []string{"I"},
		//	Value:       nil,
		//	Destination: &cliArgs.ProtoImportPaths,
		//},

		//&cli.StringSliceFlag{
		//	Name:        "enable-rule",
		//	Usage:       "Enable a rule with the given name.\nMay be specified multiple times.",
		//	Value:       nil,
		//	Destination: &cliArgs.EnabledRules,
		//},
		//
		//&cli.StringSliceFlag{
		//	Name:        "disable-rule",
		//	Usage:       "Disable a rule with the given name.\nMay be specified multiple times.",
		//	Value:       nil,
		//	Destination: &cliArgs.DisabledRules,
		//},
	}
}

// LinterConfig holds configuration for the linter.
type LinterConfig struct {
	Rules                     lint.Config `yaml:"rules,omitempty" hash:"-"`
	FormatType                string      `yaml:"format_type"`
	IgnoreCommentDisablesFlag bool        `yaml:"ignore_comment_disables_flag"`
}

// Linter runs the linter on the given proto files.
func Linter(c *CliArgs, config LinterConfig, protoImportPaths, protoFiles []string) error {
	if c.ListRulesFlag {
		return outputRules(config.FormatType)
	}

	// Pre-check if there are files to lint.
	if len(protoFiles) == 0 {
		return fmt.Errorf("no file to lint")
	}

	rules := lint.Configs{config.Rules}

	// Add configs for the enabled rules.
	rules = append(rules, lint.Config{EnabledRules: c.EnabledRules})
	rules = append(rules, lint.Config{DisabledRules: c.DisabledRules})

	// Create resolver for source files with import paths.
	importPaths := append(protoImportPaths, ".")
	sourceResolver := &protocompile.SourceResolver{
		ImportPaths: importPaths,
	}

	// Create resolver for standard imports (like google/protobuf/*.proto).
	importResolver := protocompile.ResolverFunc(func(path string) (protocompile.SearchResult, error) {
		fd, err := protoregistry.GlobalFiles.FindFileByPath(path)
		if err != nil {
			return protocompile.SearchResult{}, err
		}
		return protocompile.SearchResult{Desc: fd}, nil
	})

	// Collect errors during compilation.
	var collectedErrors []error
	rep := reporter.NewReporter(func(err reporter.ErrorWithPos) error {
		collectedErrors = append(collectedErrors, err)
		return nil // Continue on error
	}, nil)

	// Create compiler with combined resolvers.
	compiler := protocompile.Compiler{
		Resolver:       protocompile.WithStandardImports(protocompile.CompositeResolver{sourceResolver, importResolver}),
		SourceInfoMode: protocompile.SourceInfoStandard,
		Reporter:       rep,
	}

	// Compile proto files.
	compiledFiles, err := compiler.Compile(context.Background(), protoFiles...)

	// Check for collected errors first.
	if len(collectedErrors) > 0 {
		errStrings := make([]string, len(collectedErrors))
		for i, e := range collectedErrors {
			errStrings[i] = e.Error()
		}
		return errors.New(strings.Join(errStrings, "\n"))
	}

	if err != nil {
		return err
	}

	// Convert to protoreflect.FileDescriptor slice.
	fileDescriptors := make([]protoreflect.FileDescriptor, 0, len(compiledFiles))
	for _, f := range compiledFiles {
		fileDescriptors = append(fileDescriptors, f)
	}

	// Create a Linter to lint the file descriptors.
	l := lint.New(globalRules, rules, lint.Debug(c.DebugFlag), lint.IgnoreCommentDisables(config.IgnoreCommentDisablesFlag))
	results, err := l.LintProtos(fileDescriptors...)
	if err != nil {
		return err
	}

	// Determine the format for printing the results.
	// YAML format is the default.
	marshal := getOutputFormatFunc(config.FormatType)

	// Print the results.
	b, err := marshal(results)
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	filterResults := lo.Filter(results, func(item lint.Response, _ int) bool { return len(item.Problems) > 0 })
	if len(filterResults) > 0 {
		os.Exit(1)
	}

	return nil
}

var outputFormatFuncs = map[string]formatFunc{
	"yaml": yaml.Marshal,
	"yml":  yaml.Marshal,
	"json": json.Marshal,
	"github": func(i any) ([]byte, error) {
		switch v := i.(type) {
		case []lint.Response:
			return formatGitHubActionOutput(v), nil
		default:
			return json.Marshal(v)
		}
	},
}

type formatFunc func(any) ([]byte, error)

func getOutputFormatFunc(formatType string) formatFunc {
	if f, found := outputFormatFuncs[strings.ToLower(formatType)]; found {
		return f
	}
	return yaml.Marshal
}
