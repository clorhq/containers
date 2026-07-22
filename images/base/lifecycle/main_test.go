package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type fakeProcess struct {
	done       chan processResult
	signals    chan syscall.Signal
	complete   sync.Once
	mutex      sync.Mutex
	groupAlive bool
}

func newFakeProcess() *fakeProcess {
	return &fakeProcess{
		done:       make(chan processResult, 1),
		signals:    make(chan syscall.Signal, 16),
		groupAlive: true,
	}
}

func (process *fakeProcess) Done() <-chan processResult {
	return process.done
}

func (process *fakeProcess) SignalGroup(signal syscall.Signal) error {
	process.signals <- signal
	return nil
}

func (process *fakeProcess) GroupAlive() bool {
	process.mutex.Lock()
	defer process.mutex.Unlock()
	return process.groupAlive
}

func (process *fakeProcess) finish(code int) {
	process.finishWithGroup(code, false)
}

func (process *fakeProcess) finishWithGroup(code int, groupAlive bool) {
	process.complete.Do(func() {
		process.mutex.Lock()
		process.groupAlive = groupAlive
		process.mutex.Unlock()
		process.done <- processResult{code: code}
		close(process.done)
	})
}

type startedFakeProcess struct {
	spec    processSpec
	process *fakeProcess
}

type fakeRunner struct {
	started                chan startedFakeProcess
	autoFinishServiceStops bool
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{
		started:                make(chan startedFakeProcess, 16),
		autoFinishServiceStops: true,
	}
}

func (runner *fakeRunner) Start(spec processSpec) (childProcess, error) {
	process := newFakeProcess()
	if spec.role == serviceStopRole && runner.autoFinishServiceStops {
		process.finish(0)
		return process, nil
	}
	runner.started <- startedFakeProcess{spec: spec, process: process}
	return process, nil
}

type fakeTimer struct {
	duration time.Duration
	channel  chan time.Time
	stopped  bool
	mutex    sync.Mutex
}

func (timer *fakeTimer) Chan() <-chan time.Time {
	return timer.channel
}

func (timer *fakeTimer) Stop() {
	timer.mutex.Lock()
	defer timer.mutex.Unlock()
	timer.stopped = true
}

func (timer *fakeTimer) Reset(duration time.Duration) {
	timer.mutex.Lock()
	defer timer.mutex.Unlock()
	timer.duration = duration
	timer.stopped = false
}

func (timer *fakeTimer) fire() {
	timer.channel <- time.Now()
}

type fakeClock struct {
	created chan *fakeTimer
}

func newFakeClock() *fakeClock {
	return &fakeClock{created: make(chan *fakeTimer, 16)}
}

func (clock *fakeClock) NewTimer(duration time.Duration) timer {
	timer := &fakeTimer{duration: duration, channel: make(chan time.Time, 1)}
	clock.created <- timer
	return timer
}

type fakeMetadata struct {
	started     bool
	directory   string
	resetCount  int
	removeCount int
}

func (metadata *fakeMetadata) Reset() error {
	metadata.resetCount++
	return nil
}

func (metadata *fakeMetadata) RemoveReadiness() error {
	metadata.removeCount++
	return nil
}

func (metadata *fakeMetadata) SupervisionStarted() bool {
	return metadata.started
}

func (metadata *fakeMetadata) ArchiveDirectory(string) string {
	return metadata.directory
}

type coordinatorHarness struct {
	runner   *fakeRunner
	clock    *fakeClock
	metadata *fakeMetadata
	signals  chan os.Signal
	result   chan int
	states   chan lifecycleState
}

func newCoordinatorHarness(archiveScript string, supervisionStarted bool) *coordinatorHarness {
	runner := newFakeRunner()
	clock := newFakeClock()
	metadata := &fakeMetadata{started: supervisionStarted, directory: "/archive"}
	states := make(chan lifecycleState, 16)
	config := coordinatorConfig{
		workloadArgs:       []string{"entrypoint"},
		archiveScript:      []byte(archiveScript),
		spaceUser:          "space-user",
		spaceHome:          "/space-home",
		spacePath:          "/space-path",
		playwrightPath:     "/browsers",
		archiveTimeout:     defaultArchiveTimeout,
		stopGrace:          defaultStopGrace,
		serviceStopTimeout: defaultServiceStopTimeout,
		serviceStopCommand: "stop-services",
		markerPollInterval: 0,
		runner:             runner,
		clock:              clock,
		metadata:           metadata,
		stateChanged: func(state lifecycleState) {
			states <- state
		},
		logger: log.New(io.Discard, "", 0),
	}
	signals := make(chan os.Signal, 16)
	result := make(chan int, 1)
	go func() {
		result <- newCoordinator(config).Run(signals)
	}()
	return &coordinatorHarness{
		runner: runner, clock: clock, metadata: metadata,
		signals: signals, result: result, states: states,
	}
}

func awaitStartedProcess(t *testing.T, harness *coordinatorHarness, role processRole) startedFakeProcess {
	t.Helper()
	select {
	case started := <-harness.runner.started:
		if started.spec.role != role {
			t.Fatalf("started %s process, want %s", started.spec.role, role)
		}
		return started
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s process", role)
		return startedFakeProcess{}
	}
}

func assertNoStartedProcess(t *testing.T, harness *coordinatorHarness) {
	t.Helper()
	select {
	case started := <-harness.runner.started:
		t.Fatalf("unexpected %s process", started.spec.role)
	case <-time.After(25 * time.Millisecond):
	}
}

func awaitSignal(t *testing.T, process *fakeProcess, expected syscall.Signal) {
	t.Helper()
	select {
	case actual := <-process.signals:
		if actual != expected {
			t.Fatalf("received %s, want %s", actual, expected)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", expected)
	}
}

func awaitTimer(t *testing.T, harness *coordinatorHarness, duration time.Duration) *fakeTimer {
	t.Helper()
	select {
	case timer := <-harness.clock.created:
		timer.mutex.Lock()
		actualDuration := timer.duration
		stopped := timer.stopped
		timer.mutex.Unlock()
		if actualDuration != duration && stopped {
			return awaitTimer(t, harness, duration)
		}
		if actualDuration != duration {
			t.Fatalf("created timer for %s, want %s", actualDuration, duration)
		}
		return timer
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s timer", duration)
		return nil
	}
}

func awaitResult(t *testing.T, harness *coordinatorHarness, expected int) {
	t.Helper()
	select {
	case actual := <-harness.result:
		if actual != expected {
			t.Fatalf("coordinator returned %d, want %d", actual, expected)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for coordinator status %d", expected)
	}
}

func assertNoResult(t *testing.T, harness *coordinatorHarness) {
	t.Helper()
	select {
	case result := <-harness.result:
		t.Fatalf("coordinator returned early with %d", result)
	case <-time.After(25 * time.Millisecond):
	}
}

func TestArchiveIsOneShotAndReturnsScriptStatus(t *testing.T) {
	harness := newCoordinatorHarness("exit 19", true)
	workload := awaitStartedProcess(t, harness, workloadRole)

	harness.signals <- syscall.SIGUSR1
	harness.signals <- syscall.SIGUSR1
	archive := awaitStartedProcess(t, harness, archiveRole)
	if archive.spec.dir != "/archive" {
		t.Fatalf("archive directory = %q, want /archive", archive.spec.dir)
	}
	if !bytes.Equal(archive.spec.input, []byte("exit 19")) {
		t.Fatalf("archive standard input = %q, want exact decoded script", archive.spec.input)
	}
	if got := archive.spec.args[len(archive.spec.args)-1]; got != "-s" {
		t.Fatalf("archive Bash mode = %q, want -s", got)
	}
	assertNoStartedProcess(t, harness)
	archive.process.finish(19)
	awaitSignal(t, workload.process, syscall.SIGTERM)
	workload.process.finish(143)
	awaitResult(t, harness, 19)

	if harness.metadata.removeCount != 2 {
		t.Fatalf("readiness removed %d times, want 2", harness.metadata.removeCount)
	}
	if harness.metadata.resetCount != 1 {
		t.Fatalf("metadata reset %d times, want 1", harness.metadata.resetCount)
	}
	var states []lifecycleState

drainStates:
	for {
		select {
		case state := <-harness.states:
			states = append(states, state)
		default:
			break drainStates
		}
	}
	want := []lifecycleState{
		stateStarting, stateRunning, stateStopping, stateArchiving, stateStopping, stateDone,
	}
	if !reflect.DeepEqual(states, want) {
		t.Fatalf("state transitions = %v, want %v", states, want)
	}
	for {
		select {
		case signal := <-archive.process.signals:
			if signal == syscall.SIGUSR1 {
				t.Fatal("SIGUSR1 reached archive process group")
			}
		default:
			return
		}
	}
}

func TestArchiveStopsTabServicesBeforeStartingAndAllServicesAfterward(t *testing.T) {
	harness := newCoordinatorHarness("archive", true)
	harness.runner.autoFinishServiceStops = false
	workload := awaitStartedProcess(t, harness, workloadRole)

	harness.signals <- syscall.SIGUSR1
	tabStop := awaitStartedProcess(t, harness, serviceStopRole)
	if got := tabStop.spec.args[len(tabStop.spec.args)-1]; got != string(stopTabServices) {
		t.Fatalf("first service stop mode = %q, want tabs", got)
	}
	assertNoStartedProcess(t, harness)
	tabStop.process.finish(0)
	archive := awaitStartedProcess(t, harness, archiveRole)
	archive.process.finish(0)
	awaitSignal(t, workload.process, syscall.SIGTERM)
	allStop := awaitStartedProcess(t, harness, serviceStopRole)
	if got := allStop.spec.args[len(allStop.spec.args)-1]; got != string(stopAllServices) {
		t.Fatalf("final service stop mode = %q, want all", got)
	}
	workload.process.finish(143)
	assertNoResult(t, harness)
	allStop.process.finish(0)
	awaitResult(t, harness, 0)
}

func TestArchiveStopsBootstrapBeforeStarting(t *testing.T) {
	harness := newCoordinatorHarness("archive", false)
	workload := awaitStartedProcess(t, harness, workloadRole)

	harness.signals <- syscall.SIGUSR1
	awaitSignal(t, workload.process, syscall.SIGTERM)
	_ = awaitTimer(t, harness, defaultStopGrace)
	assertNoStartedProcess(t, harness)
	workload.process.finish(143)
	archive := awaitStartedProcess(t, harness, archiveRole)
	archive.process.finish(0)
	awaitResult(t, harness, 0)
}

func TestMissingArchiveScriptReturnsOne(t *testing.T) {
	harness := newCoordinatorHarness("", true)
	workload := awaitStartedProcess(t, harness, workloadRole)

	harness.signals <- syscall.SIGUSR1
	awaitSignal(t, workload.process, syscall.SIGTERM)
	assertNoStartedProcess(t, harness)
	workload.process.finish(143)
	awaitResult(t, harness, 1)
}

func TestArchiveTimeoutTerminatesThenKills(t *testing.T) {
	harness := newCoordinatorHarness("hang", true)
	workload := awaitStartedProcess(t, harness, workloadRole)
	harness.signals <- syscall.SIGUSR1
	archive := awaitStartedProcess(t, harness, archiveRole)

	archiveTimeout := awaitTimer(t, harness, defaultArchiveTimeout)
	archiveTimeout.fire()
	awaitSignal(t, archive.process, syscall.SIGTERM)
	archiveKill := awaitTimer(t, harness, defaultStopGrace)
	archiveKill.fire()
	awaitSignal(t, archive.process, syscall.SIGKILL)
	archive.process.finish(137)
	awaitSignal(t, workload.process, syscall.SIGTERM)
	workload.process.finish(143)
	awaitResult(t, harness, 124)
}

func TestArchiveTimeoutKillsDescendantsAfterLeaderExits(t *testing.T) {
	harness := newCoordinatorHarness("hang", true)
	workload := awaitStartedProcess(t, harness, workloadRole)
	harness.signals <- syscall.SIGUSR1
	archive := awaitStartedProcess(t, harness, archiveRole)

	archiveTimeout := awaitTimer(t, harness, defaultArchiveTimeout)
	archiveTimeout.fire()
	awaitSignal(t, archive.process, syscall.SIGTERM)
	archiveKill := awaitTimer(t, harness, defaultStopGrace)
	archive.process.finishWithGroup(143, true)
	assertNoResult(t, harness)
	archiveKill.fire()
	awaitSignal(t, archive.process, syscall.SIGKILL)
	awaitSignal(t, archive.process, syscall.SIGTERM)
	awaitSignal(t, workload.process, syscall.SIGTERM)
	workload.process.finish(143)
	awaitResult(t, harness, 124)
}

func TestTerminationBeforeArchiveUsesSignalStatus(t *testing.T) {
	harness := newCoordinatorHarness("archive", true)
	workload := awaitStartedProcess(t, harness, workloadRole)

	harness.signals <- syscall.SIGTERM
	awaitSignal(t, workload.process, syscall.SIGTERM)
	workload.process.finish(143)
	awaitResult(t, harness, 143)
	assertNoStartedProcess(t, harness)
}

func TestInterruptDuringArchiveForwardsToBothGroups(t *testing.T) {
	harness := newCoordinatorHarness("archive", true)
	workload := awaitStartedProcess(t, harness, workloadRole)
	harness.signals <- syscall.SIGUSR1
	archive := awaitStartedProcess(t, harness, archiveRole)

	harness.signals <- syscall.SIGINT
	awaitSignal(t, archive.process, syscall.SIGINT)
	awaitSignal(t, workload.process, syscall.SIGINT)
	archive.process.finish(130)
	workload.process.finish(130)
	awaitResult(t, harness, 130)
}

func TestWorkloadExitDoesNotCutArchiveShort(t *testing.T) {
	harness := newCoordinatorHarness("archive", true)
	workload := awaitStartedProcess(t, harness, workloadRole)
	harness.signals <- syscall.SIGUSR1
	archive := awaitStartedProcess(t, harness, archiveRole)

	workload.process.finish(42)
	assertNoResult(t, harness)
	archive.process.finish(0)
	awaitResult(t, harness, 0)
}

func TestSystemRunnerSignalsTheWholeProcessGroup(t *testing.T) {
	temporaryDirectory := t.TempDir()
	childSignalFile := filepath.Join(temporaryDirectory, "child-signaled")
	childReadyFile := filepath.Join(temporaryDirectory, "child-ready")
	parentReadyFile := filepath.Join(temporaryDirectory, "parent-ready")
	script := `
set -eu
bash -c 'trap '\''printf signaled >"$1"; exit 0'\'' TERM; printf ready >"$2"; while :; do sleep 1; done' _ "$1" "$2" &
child=$!
trap 'wait "$child" 2>/dev/null || true; exit 0' TERM
printf ready >"$3"
wait "$child"
`
	process, err := (systemRunner{}).Start(processSpec{
		role: workloadRole,
		args: []string{"bash", "-c", script, "_", childSignalFile, childReadyFile, parentReadyFile},
	})
	if err != nil {
		t.Fatalf("start process group: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		if _, childErr := os.Stat(childReadyFile); childErr == nil {
			if _, parentErr := os.Stat(parentReadyFile); parentErr == nil {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("nested process did not become ready")
		}
		time.Sleep(time.Millisecond)
	}
	if err := process.SignalGroup(syscall.SIGTERM); err != nil {
		t.Fatalf("signal process group: %v", err)
	}
	select {
	case <-process.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("process group leader did not exit")
	}
	if _, err := os.Stat(childSignalFile); err != nil {
		t.Fatalf("process-group child did not receive SIGTERM: %v", err)
	}
}

func TestExitCodeMapping(t *testing.T) {
	if code := exitCode(nil); code != 0 {
		t.Fatalf("nil error maps to %d, want 0", code)
	}
	command := exec.Command("bash", "-c", "exit 37")
	if code := exitCode(command.Run()); code != 37 {
		t.Fatalf("script failure maps to %d, want 37", code)
	}
}

func TestInvalidLifecycleEncodingFailsBeforeWorkloadStarts(t *testing.T) {
	runner := newFakeRunner()
	metadata := &fakeMetadata{}
	config := coordinatorConfig{
		workloadArgs:       []string{"entrypoint"},
		configurationError: errors.New("invalid lifecycle Base64"),
		runner:             runner,
		clock:              newFakeClock(),
		metadata:           metadata,
		logger:             log.New(io.Discard, "", 0),
	}
	if code := newCoordinator(config).Run(make(chan os.Signal)); code != 1 {
		t.Fatalf("status = %d, want 1", code)
	}
	select {
	case started := <-runner.started:
		t.Fatalf("invalid configuration started %s", started.spec.role)
	default:
	}
	if metadata.removeCount != 1 {
		t.Fatalf("readiness removed %d times, want 1", metadata.removeCount)
	}
}

func TestArchiveDirectoryFallsBackToSpaceHome(t *testing.T) {
	temporaryDirectory := t.TempDir()
	metadataDirectory := filepath.Join(temporaryDirectory, "metadata")
	archiveDirectory := filepath.Join(temporaryDirectory, "repository with spaces")
	if err := os.MkdirAll(metadataDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(archiveDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	metadata := fileMetadata{lifecycleDirectory: metadataDirectory}
	home := filepath.Join(temporaryDirectory, "home")

	if actual := metadata.ArchiveDirectory(home); actual != home {
		t.Fatalf("missing metadata directory = %q, want %q", actual, home)
	}
	metadataFile := filepath.Join(metadataDirectory, archiveDirectoryFilename)
	if err := os.WriteFile(metadataFile, []byte(archiveDirectory), 0o644); err != nil {
		t.Fatal(err)
	}
	if actual := metadata.ArchiveDirectory(home); actual != archiveDirectory {
		t.Fatalf("metadata directory = %q, want %q", actual, archiveDirectory)
	}
	if err := os.WriteFile(metadataFile, []byte(filepath.Join(temporaryDirectory, "missing")), 0o644); err != nil {
		t.Fatal(err)
	}
	if actual := metadata.ArchiveDirectory(home); actual != home {
		t.Fatalf("invalid metadata directory = %q, want %q", actual, home)
	}
}

func TestDecodeScriptEnvironmentIsBase64OnlyAndByteExact(t *testing.T) {
	const name = "CLOR_TEST_SCRIPT_BASE64"
	t.Setenv("CLOR_TEST_SCRIPT", "legacy script")
	original := []byte("first line\r\nsecond line\n\n")
	t.Setenv(name, base64.StdEncoding.EncodeToString(original))

	actual, err := decodeScriptEnvironment(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, original) {
		t.Fatalf("decoded script = %q, want %q", actual, original)
	}

	t.Setenv(name, "not valid base64!")
	if _, err := decodeScriptEnvironment(name); err == nil {
		t.Fatal("invalid Base64 script succeeded")
	}

	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}
	actual, err = decodeScriptEnvironment(name)
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) != 0 {
		t.Fatalf("raw compatibility variable was read: %q", actual)
	}
}

func TestArchiveEnvironmentExcludesLifecycleAndRuntimeSecrets(t *testing.T) {
	t.Setenv("CLOR_SETUP_SCRIPT_BASE64", base64.StdEncoding.EncodeToString([]byte("setup")))
	t.Setenv("CLOR_ARCHIVE_SCRIPT_BASE64", base64.StdEncoding.EncodeToString([]byte("archive")))
	t.Setenv("CLOR_API_KEY", "api-secret")
	t.Setenv("CLOR_TAB_EDITOR_WEBTUI_SECRET", "web-secret")
	t.Setenv("CLOR_TAB_EDITOR_DIRECT_TOKEN", "direct-secret")
	t.Setenv("AUTHORED_MULTILINE", "first\r\nsecond\n")

	config := configFromEnvironment([]string{"entrypoint"})
	values := make(map[string]string)
	for _, variable := range config.archiveEnv {
		name, value, _ := strings.Cut(variable, "=")
		values[name] = value
	}
	for _, name := range []string{
		"CLOR_SETUP_SCRIPT_BASE64", "CLOR_ARCHIVE_SCRIPT_BASE64", "CLOR_API_KEY",
		"CLOR_TAB_EDITOR_WEBTUI_SECRET", "CLOR_TAB_EDITOR_DIRECT_TOKEN",
	} {
		if _, exists := values[name]; exists {
			t.Fatalf("archive environment retained %s", name)
		}
	}
	if values["AUTHORED_MULTILINE"] != "first\r\nsecond\n" {
		t.Fatalf("authored multiline environment changed to %q", values["AUTHORED_MULTILINE"])
	}
}
