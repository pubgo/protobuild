package linters

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"os"
	"strings"
	"sync"

	"github.com/googleapis/api-linter/lint"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pubgo/protobuild/internal/typex"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type CliArgs struct {
	//FormatType string
	//ProtoImportPaths          []string
	EnabledRules  []string
	DisabledRules []string
	ListRulesFlag bool
	DebugFlag     bool
	//IgnoreCommentDisablesFlag bool
}

func NewCli() (*CliArgs, typex.Flags) {
	var cliArgs CliArgs

	return &cliArgs, typex.Flags{
		//&cli.BoolFlag{
		//	Name:        "ignore-comment-disables",
		//	Usage:       "If set to true, disable comments will be ignored.\nThis is helpful when strict enforcement of AIPs are necessary and\nproto definitions should not be able to disable checks.",
		//	Value:       false,
		//	Destination: &cliArgs.IgnoreCommentDisablesFlag,
		//},

		&cli.BoolFlag{
			Name:        "debug",
			Usage:       "Run in debug mode. Panics will print stack.",
			Value:       false,
			Destination: &cliArgs.DebugFlag,
		},

		&cli.BoolFlag{
			Name:        "list-rules",
			Usage:       "Print the rules and exit.  Honors the output-format flag.",
			Value:       false,
			Destination: &cliArgs.ListRulesFlag,
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

type LinterConfig struct {
	Rules                     lint.Config `yaml:"rules,omitempty" hash:"-"`
	FormatType                string      `yaml:"format_type"`
	IgnoreCommentDisablesFlag bool        `yaml:"ignore_comment_disables_flag"`
}

func Linter(c *CliArgs, config LinterConfig, protoImportPaths []string, protoFiles []string) error {
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

	var errorsWithPos []protoparse.ErrorWithPos
	var lock sync.Mutex
	// Parse proto files into `protoreflect` file descriptors.
	p := protoparse.Parser{
		ImportPaths:           append(protoImportPaths, "."),
		IncludeSourceCodeInfo: true,
		ErrorReporter: func(errorWithPos protoparse.ErrorWithPos) error {
			// Protoparse isn't concurrent right now but just to be safe for the future.
			lock.Lock()
			errorsWithPos = append(errorsWithPos, errorWithPos)
			lock.Unlock()
			// Continue parsing. The error returned will be protoparse.ErrInvalidSource.
			return nil
		},
	}

	var err error
	// Resolve file absolute paths to relative ones.
	// Using supplied import paths first.
	if len(protoImportPaths) > 0 {
		protoFiles, err = protoparse.ResolveFilenames(protoImportPaths, protoFiles...)
		if err != nil {
			return err
		}
	}
	// Then resolve again against ".", the local directory.
	// This is necessary because ResolveFilenames won't resolve a path if it
	// relative to *at least one* of the given import paths, which can result
	// in duplicate file parsing and compilation errors, as seen in #1465 and
	// #1471. So we resolve against local (default) and flag specified import
	// paths separately.
	protoFiles, err = protoparse.ResolveFilenames([]string{"."}, protoFiles...)
	if err != nil {
		return err
	}

	fd, err := p.ParseFiles(protoFiles...)
	if err != nil {
		if err == protoparse.ErrInvalidSource {
			if len(errorsWithPos) == 0 {
				return errors.New("got protoparse.ErrInvalidSource but no ErrorWithPos errors")
			}
			// TODO: There's multiple ways to deal with this but this prints all the errors at least
			errStrings := make([]string, len(errorsWithPos))
			for i, errorWithPos := range errorsWithPos {
				errStrings[i] = errorWithPos.Error()
			}
			return errors.New(strings.Join(errStrings, "\n"))
		}
		return err
	}

	// Create a Linter to lint the file descriptors.
	l := lint.New(globalRules, rules, lint.Debug(c.DebugFlag), lint.IgnoreCommentDisables(config.IgnoreCommentDisablesFlag))
	results, err := l.LintProtos(fd...)
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

	filterResults := lo.Filter(results, func(item lint.Response, index int) bool { return len(item.Problems) > 0 })
	if len(filterResults) > 0 {
		os.Exit(1)
	}

	return nil
}

var outputFormatFuncs = map[string]formatFunc{
	"yaml": yaml.Marshal,
	"yml":  yaml.Marshal,
	"json": json.Marshal,
	"github": func(i interface{}) ([]byte, error) {
		switch v := i.(type) {
		case []lint.Response:
			return formatGitHubActionOutput(v), nil
		default:
			return json.Marshal(v)
		}
	},
}

type formatFunc func(interface{}) ([]byte, error)

func getOutputFormatFunc(formatType string) formatFunc {
	if f, found := outputFormatFuncs[strings.ToLower(formatType)]; found {
		return f
	}
	return yaml.Marshal
}
