package linters

import (
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"path/filepath"
	"strings"

	"github.com/googleapis/api-linter/lint"
)

// formatGitHubActionOutput returns lint errors in GitHub actions format.
func formatGitHubActionOutput(responses []lint.Response) []byte {
	var buf bytes.Buffer
	for _, response := range responses {
		for _, problem := range response.Problems {
			// lint example:
			// ::error file={name},line={line},endLine={endLine},title={title}::{message}
			// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-error-message

			fmt.Println(lo.Must(filepath.Abs(response.FilePath)))
			fmt.Fprintf(&buf, "::error file=%s", response.FilePath)
			if problem.Location != nil {
				// Some findings are *line level* and only have start positions but no
				// starting column. Construct a switch fallthrough to emit as many of
				// the location indicators are included.
				switch len(problem.Location.Span) {
				case 4:
					fmt.Fprintf(&buf, ",endColumn=%d", problem.Location.Span[3])
					fallthrough
				case 3:
					fmt.Fprintf(&buf, ",endLine=%d", problem.Location.Span[2])
					fallthrough
				case 2:
					fmt.Fprintf(&buf, ",col=%d", problem.Location.Span[1])
					fallthrough
				case 1:
					fmt.Fprintf(&buf, ",line=%d", problem.Location.Span[0])
				}
			}

			// GitHub uses :: as control characters (which are also used to delimit
			// Linter rules. In order to prevent confusion, replace the double colon
			// with two Armenian full stops which are indistinguishable to my eye.
			runeThatLooksLikeTwoColonsButIsActuallyTwoArmenianFullStops := "։։"
			title := strings.ReplaceAll(string(problem.RuleID), "::", runeThatLooksLikeTwoColonsButIsActuallyTwoArmenianFullStops)
			message := strings.ReplaceAll(problem.Message, "\n", "\\n")
			uri := problem.GetRuleURI()
			if uri != "" {
				message += "\\n\\n" + uri
			}
			fmt.Fprintf(&buf, ",title=%s::%s\n", title, message)
		}
	}

	return buf.Bytes()
}
