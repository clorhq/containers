package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const integrationWorkload = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

STATE="${CLOR_LIFECYCLE_TEST_STATE}"
LIFECYCLE="${CLOR_LIFECYCLE_TEST_METADATA}"
MODE="${CLOR_LIFECYCLE_TEST_WORKLOAD_MODE}"
mkdir --parents "${LIFECYCLE}" "${STATE}/workdir"
printf '%s' "$$" >"${STATE}/workload.pid"

bash -c '
    trap '\''printf "term\n" >>"$1/workload-child-signals"; exit 0'\'' TERM
    trap '\''printf "int\n" >>"$1/workload-child-signals"; exit 0'\'' INT
    printf "%s" "$$" >"$1/workload-child.pid"
    while :; do sleep 1; done
' _ "${STATE}" &
CHILD="$!"

stop_term() {
    printf 'term\n' >>"${STATE}/workload-signals"
    wait "${CHILD}" 2>/dev/null || true
    exit 0
}
stop_int() {
    printf 'int\n' >>"${STATE}/workload-signals"
    wait "${CHILD}" 2>/dev/null || true
    exit 0
}
trap stop_term TERM
trap stop_int INT

if [[ "${MODE}" == "bootstrap" ]]; then
    : >"${STATE}/bootstrap-started"
else
    printf '%s' "${STATE}/workdir" >"${LIFECYCLE}/archive-directory"
    : >"${LIFECYCLE}/workload-started"
    : >"${CLOR_LIFECYCLE_TEST_READY}"
fi

if [[ "${MODE}" == "exit" ]]; then
    sleep 0.08
    kill -TERM "${CHILD}" 2>/dev/null || true
    wait "${CHILD}" 2>/dev/null || true
    exit 42
fi
wait "${CHILD}"
`

type integrationProcess struct {
	command *exec.Cmd
	waited  bool
}

func (process *integrationProcess) stop() {
	if process.waited || process.command.Process == nil {
		return
	}
	_ = process.command.Process.Kill()
	_ = process.command.Wait()
	process.waited = true
}

func TestCoordinatorIntegrationHelper(t *testing.T) {
	if os.Getenv("CLOR_LIFECYCLE_INTEGRATION_HELPER") != "1" {
		return
	}

	config := configFromEnvironment([]string{"bash", os.Getenv("CLOR_LIFECYCLE_TEST_ENTRYPOINT")})
	config.archiveTimeout = 180 * time.Millisecond
	config.stopGrace = 120 * time.Millisecond
	config.serviceStopTimeout = 120 * time.Millisecond
	config.serviceStopCommand = os.Getenv("CLOR_LIFECYCLE_TEST_SERVICE_STOP")
	config.markerPollInterval = 5 * time.Millisecond
	config.metadata = fileMetadata{
		lifecycleDirectory: os.Getenv("CLOR_LIFECYCLE_TEST_METADATA"),
		readyFile:          os.Getenv("CLOR_LIFECYCLE_TEST_READY"),
	}
	signals := make(chan os.Signal, 16)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT)
	os.Exit(newCoordinator(config).Run(signals))
}

func TestDumbInitIntegration(t *testing.T) {
	if os.Getenv("CLOR_LIFECYCLE_INTEGRATION") != "1" {
		t.Skip("set CLOR_LIFECYCLE_INTEGRATION=1 to run dumb-init integration tests")
	}
	if _, err := exec.LookPath("gosu"); err != nil && os.Getenv("CLOR_LIFECYCLE_FAKE_GOSU") == "1" {
		fakeBin := t.TempDir()
		fakeGosu := filepath.Join(fakeBin, "gosu")
		if err := os.WriteFile(fakeGosu, []byte("#!/bin/bash\nshift\nexec \"$@\"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", fakeBin+":"+os.Getenv("PATH"))
	}
	for _, command := range []string{"dumb-init", "gosu", "bash"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Fatalf("required integration command %q is unavailable: %v", command, err)
		}
	}

	t.Run("successful one-shot archive after supervision", func(t *testing.T) {
		process, state := startIntegrationProcess(t, "running", archiveScript("success"))
		defer process.stop()
		waitForFile(t, filepath.Join(state, "workload-child.pid"))
		signalIntegrationProcess(t, process, syscall.SIGUSR1)
		signalIntegrationProcess(t, process, syscall.SIGUSR1)
		if code := waitForIntegrationProcess(t, process); code != 0 {
			t.Fatalf("status = %d, want 0", code)
		}
		starts, err := os.ReadFile(filepath.Join(state, "archive-starts"))
		if err != nil {
			t.Fatal(err)
		}
		if lines := strings.Count(string(starts), "start\n"); lines != 1 {
			t.Fatalf("archive starts = %d, want 1", lines)
		}
		workingDirectory, err := os.ReadFile(filepath.Join(state, "archive-pwd"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := strings.TrimSpace(string(workingDirectory)), filepath.Join(state, "workdir"); got != want {
			t.Fatalf("archive working directory = %q, want %q", got, want)
		}
		serviceStops, err := os.ReadFile(filepath.Join(state, "service-stops"))
		if err != nil {
			t.Fatal(err)
		}
		if got, want := string(serviceStops), "tabs\nall\n"; got != want {
			t.Fatalf("service stop order = %q, want %q", got, want)
		}
	})

	t.Run("script failure and missing script", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			script string
			code   int
		}{
			{name: "failure", script: archiveScript("failure"), code: 23},
			{name: "missing", script: "", code: 1},
		} {
			t.Run(test.name, func(t *testing.T) {
				process, state := startIntegrationProcess(t, "running", test.script)
				defer process.stop()
				waitForFile(t, filepath.Join(state, "workload.pid"))
				signalIntegrationProcess(t, process, syscall.SIGUSR1)
				if code := waitForIntegrationProcess(t, process); code != test.code {
					t.Fatalf("status = %d, want %d", code, test.code)
				}
			})
		}
	})

	t.Run("timeout escalates after termination", func(t *testing.T) {
		process, state := startIntegrationProcess(t, "running", archiveScript("hang"))
		defer process.stop()
		waitForFile(t, filepath.Join(state, "workload.pid"))
		signalIntegrationProcess(t, process, syscall.SIGUSR1)
		waitForFile(t, filepath.Join(state, "archive-starts"))
		started := time.Now()
		if code := waitForIntegrationProcess(t, process); code != 124 {
			t.Fatalf("status = %d, want 124", code)
		}
		if elapsed := time.Since(started); elapsed < 100*time.Millisecond {
			t.Fatalf("timeout escalation lasted %s, want at least the kill grace", elapsed)
		}
		assertFileContains(t, filepath.Join(state, "archive-signals"), "term")
	})

	t.Run("forced stop before archive", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			signal syscall.Signal
			code   int
		}{
			{name: "term", signal: syscall.SIGTERM, code: 143},
			{name: "int", signal: syscall.SIGINT, code: 130},
		} {
			t.Run(test.name, func(t *testing.T) {
				process, state := startIntegrationProcess(t, "running", archiveScript("success"))
				defer process.stop()
				waitForFile(t, filepath.Join(state, "workload-child.pid"))
				signalIntegrationProcess(t, process, test.signal)
				if code := waitForIntegrationProcess(t, process); code != test.code {
					t.Fatalf("status = %d, want %d", code, test.code)
				}
				if _, err := os.Stat(filepath.Join(state, "archive-starts")); !os.IsNotExist(err) {
					t.Fatalf("archive unexpectedly started: %v", err)
				}
			})
		}
	})

	t.Run("forced stop during archive", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			signal syscall.Signal
			code   int
			log    string
		}{
			{name: "term", signal: syscall.SIGTERM, code: 143, log: "term"},
			{name: "int", signal: syscall.SIGINT, code: 130, log: "int"},
		} {
			t.Run(test.name, func(t *testing.T) {
				process, state := startIntegrationProcess(t, "running", archiveScript("wait"))
				defer process.stop()
				waitForFile(t, filepath.Join(state, "workload-child.pid"))
				signalIntegrationProcess(t, process, syscall.SIGUSR1)
				waitForFile(t, filepath.Join(state, "archive-child.pid"))
				signalIntegrationProcess(t, process, test.signal)
				if code := waitForIntegrationProcess(t, process); code != test.code {
					t.Fatalf("status = %d, want %d", code, test.code)
				}
				assertFileContains(t, filepath.Join(state, "archive-signals"), test.log)
				assertFileContains(t, filepath.Join(state, "workload-signals"), test.log)
			})
		}
	})

	t.Run("bootstrap is stopped before archive", func(t *testing.T) {
		process, state := startIntegrationProcess(t, "bootstrap", archiveScript("bootstrap"))
		defer process.stop()
		waitForFile(t, filepath.Join(state, "bootstrap-started"))
		signalIntegrationProcess(t, process, syscall.SIGUSR1)
		if code := waitForIntegrationProcess(t, process); code != 0 {
			t.Fatalf("status = %d, want 0", code)
		}
		assertFileContains(t, filepath.Join(state, "workload-signals"), "term")
		assertFileContains(t, filepath.Join(state, "bootstrap-order"), "stopped")
	})

	t.Run("workload exit waits for archive", func(t *testing.T) {
		process, state := startIntegrationProcess(t, "exit", archiveScript("slow-success"))
		defer process.stop()
		waitForFile(t, filepath.Join(state, "workload-child.pid"))
		signalIntegrationProcess(t, process, syscall.SIGUSR1)
		waitForFile(t, filepath.Join(state, "archive-starts"))
		time.Sleep(140 * time.Millisecond)
		if err := process.command.Process.Signal(syscall.Signal(0)); err != nil {
			t.Fatalf("dumb-init exited before archive completed: %v", err)
		}
		if code := waitForIntegrationProcess(t, process); code != 124 {
			// The test archive intentionally exceeds the shortened integration
			// timeout, proving the workload exit does not end the container.
			t.Fatalf("status = %d, want timeout status 124", code)
		}
	})
}

func startIntegrationProcess(t *testing.T, workloadMode, script string) (*integrationProcess, string) {
	t.Helper()
	state := t.TempDir()
	metadataDirectory := filepath.Join(state, "metadata")
	readyFile := filepath.Join(state, "ready")
	entrypoint := filepath.Join(state, "entrypoint")
	serviceStop := filepath.Join(state, "stop-services")
	if err := os.WriteFile(entrypoint, []byte(integrationWorkload), 0o755); err != nil {
		t.Fatal(err)
	}
	serviceStopScript := `#!/bin/bash
set -o errexit
set -o nounset
printf '%s\n' "$1" >>"${CLOR_LIFECYCLE_TEST_STATE}/service-stops"
`
	if err := os.WriteFile(serviceStop, []byte(serviceStopScript), 0o755); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(
		"dumb-init", "--single-child", "--",
		os.Args[0], "-test.run=^TestCoordinatorIntegrationHelper$",
	)
	command.Env = append(os.Environ(),
		"CLOR_LIFECYCLE_INTEGRATION_HELPER=1",
		"CLOR_LIFECYCLE_TEST_ENTRYPOINT="+entrypoint,
		"CLOR_LIFECYCLE_TEST_SERVICE_STOP="+serviceStop,
		"CLOR_LIFECYCLE_TEST_METADATA="+metadataDirectory,
		"CLOR_LIFECYCLE_TEST_READY="+readyFile,
		"CLOR_LIFECYCLE_TEST_STATE="+state,
		"CLOR_LIFECYCLE_TEST_WORKLOAD_MODE="+workloadMode,
		"CLOR_ARCHIVE_SCRIPT_BASE64="+base64.StdEncoding.EncodeToString([]byte(script)),
		"CLOR_SPACE_USER=root",
		"CLOR_SPACE_HOME="+state,
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	return &integrationProcess{command: command}, state
}

func signalIntegrationProcess(t *testing.T, process *integrationProcess, signal syscall.Signal) {
	t.Helper()
	if err := process.command.Process.Signal(signal); err != nil {
		t.Fatalf("send %s to dumb-init: %v", signal, err)
	}
}

func waitForIntegrationProcess(t *testing.T, process *integrationProcess) int {
	t.Helper()
	result := make(chan error, 1)
	go func() {
		result <- process.command.Wait()
	}()
	select {
	case err := <-result:
		process.waited = true
		if err == nil {
			return 0
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		t.Fatalf("wait for dumb-init: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for dumb-init")
	}
	return -1
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
}

func assertFileContains(t *testing.T, path, expected string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), expected) {
		t.Fatalf("%s does not contain %q: %q", path, expected, string(contents))
	}
}

func archiveScript(mode string) string {
	state := `"${CLOR_LIFECYCLE_TEST_STATE}"`
	prefix := fmt.Sprintf(`
set -o nounset
STATE=%s
printf 'start\n' >>"${STATE}/archive-starts"
pwd >"${STATE}/archive-pwd"
`, state)
	switch mode {
	case "success":
		return prefix + `sleep 0.08`
	case "slow-success":
		return prefix + `sleep 0.35`
	case "failure":
		return prefix + `exit 23`
	case "bootstrap":
		return prefix + `
if grep --quiet term "${STATE}/workload-signals"; then
    printf 'stopped\n' >"${STATE}/bootstrap-order"
else
    printf 'concurrent\n' >"${STATE}/bootstrap-order"
fi
`
	case "hang":
		return prefix + `
trap 'printf "term\n" >>"${STATE}/archive-signals"' TERM
bash -c 'trap "" TERM; while :; do sleep 1; done' &
CHILD="$!"
printf '%s' "${CHILD}" >"${STATE}/archive-child.pid"
while :; do wait "${CHILD}" || true; done
`
	case "wait":
		return prefix + `
CHILD=""
stop_term() { printf 'term\n' >>"${STATE}/archive-signals"; wait "${CHILD}" 2>/dev/null || true; exit 0; }
stop_int() { printf 'int\n' >>"${STATE}/archive-signals"; wait "${CHILD}" 2>/dev/null || true; exit 0; }
trap stop_term TERM
trap stop_int INT
bash -c '
    trap '\''exit 0'\'' TERM
    trap '\''exit 0'\'' INT
    while :; do sleep 1; done
' &
CHILD="$!"
printf '%s' "${CHILD}" >"${STATE}/archive-child.pid"
wait "${CHILD}"
`
	default:
		panic("unknown archive script mode " + strconv.Quote(mode))
	}
}
