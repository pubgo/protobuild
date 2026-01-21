// Package shutil provides shell command utilities.
package shutil

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
)

// Run executes a shell command and returns the output.
func Run(args ...string) (string, error) {
	b := bytes.NewBufferString("")

	cmd := Shell(args...)
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, strings.Join(args, " "))
	}

	return strings.TrimSpace(b.String()), nil
}

// MustRun executes a shell command and panics on error.
func MustRun(args ...string) string {
	return assert.Must1(Run(args...))
}

// GoModGraph returns the output of 'go mod graph'.
func GoModGraph() (string, error) {
	return Run("go", "mod", "graph")
}

// GoList returns the output of 'go list ./...'.
func GoList() (string, error) {
	return Run("go", "list", "./...")
}

// Shell creates an exec.Cmd for the given shell command.
func Shell(args ...string) *exec.Cmd {
	shell := strings.Join(args, " ")
	cmd := exec.Command("/bin/sh", "-c", shell)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}

// IsHelp checks if the last argument is --help or -h.
func IsHelp() bool {
	help := strings.TrimSpace(os.Args[len(os.Args)-1])
	if strings.HasSuffix(help, "--help") || strings.HasSuffix(help, "-h") {
		return true
	}
	return false
}
