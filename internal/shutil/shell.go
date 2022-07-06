package shutil

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/pubgo/funk"
)

func Run(args ...string) (string, error) {
	b := bytes.NewBufferString("")

	cmd := Shell(args...)
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return "", funk.Wrap(err, strings.Join(args, " "))
	}

	return strings.TrimSpace(b.String()), nil
}

func MustRun(args ...string) string {
	return funk.Must1(Run(args...))
}

func GoModGraph() (string, error) {
	return Run("go", "mod", "graph")
}

func GoList() (string, error) {
	return Run("go", "list", "./...")
}

func Shell(args ...string) *exec.Cmd {
	var shell = strings.Join(args, " ")
	cmd := exec.Command("/bin/sh", "-c", shell)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd
}
