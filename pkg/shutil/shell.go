package shutil

import (
	"bytes"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
)

func Run(args ...string) (string, error) {
	b := bytes.NewBufferString("")

	cmd := Shell(args...)
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, strings.Join(args, " "))
	}

	return strings.TrimSpace(b.String()), nil
}

func MustRun(args ...string) string {
	return assert.Must1(Run(args...))
}

func GoModGraph() (string, error) {
	return Run("go", "mod", "graph")
}

func GoList() (string, error) {
	return Run("go", "list", "./...")
}

func Shell(args ...string) *exec.Cmd {
	shell := strings.Join(args, " ")
	slog.Info(shell)
	cmd := exec.Command("/bin/sh", "-c", shell)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}

func IsHelp() bool {
	help := strings.TrimSpace(os.Args[len(os.Args)-1])
	if strings.HasSuffix(help, "--help") || strings.HasSuffix(help, "-h") {
		return true
	}
	return false
}
