package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestPendingHelper runs the command line of GLD_TEST_PENDING_ARGS, for TestPendingExitCode.
func TestPendingHelper(t *testing.T) {
	args := os.Getenv("GLD_TEST_PENDING_ARGS")
	if args == "" {
		t.Skip("helper process only")
	}
	rootCmd.SetArgs(strings.Split(args, "\n"))
	Execute()
}

// 013-AC12: exit 0 when a line is printed, 1 when none, 2 when the instance is gone or a path is
// missing: a Monitor that loops on 1 must stop.
func TestPendingExitCode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	open := `{version: 2, file: "a.md", annotations: [{id: "e1", exact: null, file_hash: "", thread: [{by: "reader", at: null, text: "fix"}]}]}`
	done := `{version: 2, file: "b.md", annotations: [{id: "e2", exact: null, file_hash: "", status: "done", thread: [{by: "reader", at: null, text: "fix"}]}]}`
	for name, body := range map[string]string{"a.annotations.json5": open, "b.annotations.json5": done} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// No registry file for this PID: routing off, every open entry is its own.
	pid := strconv.Itoa(os.Getpid())
	exited := exec.Command(os.Args[0], "-test.run=^$") //nolint:gosec // the test binary itself
	if err := exited.Run(); err != nil {
		t.Fatal(err)
	}
	dead := strconv.Itoa(exited.Process.Pid)
	for _, tc := range []struct {
		file string
		pid  string
		code int
		out  string
	}{
		{"a.annotations.json5", pid, 0, "\te1\n"},
		{"b.annotations.json5", pid, 1, ""},
		{"a.annotations.json5", dead, 2, ""},
		{"missing", pid, 2, ""},
		{"--port=x", pid, 2, ""},
	} {
		arg := filepath.Join(dir, tc.file)
		if strings.HasPrefix(tc.file, "--") {
			arg = tc.file
		}
		home := t.TempDir()
		cmd := exec.Command(os.Args[0], "-test.run=^TestPendingHelper$") //nolint:gosec // the test binary itself
		cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME="+home, "LOCALAPPDATA="+home,
			"GLD_TEST_PENDING_ARGS="+strings.Join([]string{"pending", "--port", "6419", "--pid", tc.pid, arg}, "\n"))
		out, err := cmd.Output()
		code := 0
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			code = exit.ExitCode()
		} else if err != nil {
			t.Fatal(err)
		}
		if code != tc.code || !strings.Contains(string(out), tc.out) {
			t.Errorf("%s, pid %s: expected exit %d and %q, got %d %q", tc.file, tc.pid, tc.code, tc.out, code, out)
		}
	}
}
