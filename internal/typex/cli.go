// Package typex provides common type aliases for CLI.
package typex

import (
	"github.com/pubgo/redant"
)

// Common type aliases for CLI components.
type (
	// Strs is an alias for string slice.
	Strs = []string
	// Options is an alias for redant.OptionSet.
	Options = redant.OptionSet
	// Command is an alias for redant.Command.
	Command = redant.Command
	// Commands is an alias for redant.Command slice.
	Commands = []*redant.Command
)
