package protobuild

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const exitOnErrorChildEnv = "PROTOBUILD_EXIT_ON_ERROR_CHILD"

// TestExitOnErrorLogsWithoutStack verifies the top-level CLI error path
// logs one line and exits with code 1 WITHOUT dumping the goroutine stack
// (funk assert.Exit always calls debug.PrintStack, which reads like a crash
// for ordinary operational errors such as "protobuf.yaml 已存在").
func TestExitOnErrorLogsWithoutStack(t *testing.T) {
	if os.Getenv(exitOnErrorChildEnv) == "1" {
		ExitOnError(errors.New("exit-on-error-marker: 配置文件 protobuf.yaml 已存在，使用 --force 覆盖"))
		t.Fatal("ExitOnError must not return on error") // unreachable: os.Exit(1)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestExitOnErrorLogsWithoutStack$")
	cmd.Env = append(os.Environ(), exitOnErrorChildEnv+"=1")
	out, err := cmd.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		t.Fatalf("want exit code 1, got err=%v, output:\n%s", err, out)
	}
	got := string(out)
	if !strings.Contains(got, "exit-on-error-marker") {
		t.Errorf("error was not logged, output:\n%s", got)
	}
	if strings.Contains(got, "goroutine") {
		t.Errorf("goroutine stack dumped; want log line only, output:\n%s", got)
	}
}

func TestExitOnErrorNil(t *testing.T) {
	ExitOnError(nil) // must return normally, no exit
}
