package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	defaultSpaceUser          = "user"
	defaultSpaceHome          = "/home/user"
	defaultLifecycleDir       = "/run/clor/lifecycle"
	defaultReadyFile          = "/run/clor/ready"
	defaultArchiveTimeout     = 5 * time.Minute
	defaultStopGrace          = 10 * time.Second
	defaultServiceStopTimeout = 20 * time.Second
	defaultMarkerPoll         = 50 * time.Millisecond
	defaultServiceStopCommand = "/usr/local/lib/clor/stop-services"
	archiveDirectoryFilename  = "archive-directory"
	workloadStartedFilename   = "workload-started"
)

type lifecycleState string

const (
	stateStarting  lifecycleState = "starting"
	stateRunning   lifecycleState = "running"
	stateArchiving lifecycleState = "archiving"
	stateStopping  lifecycleState = "stopping"
	stateDone      lifecycleState = "done"
)

type processRole string

const (
	workloadRole    processRole = "workload"
	archiveRole     processRole = "archive"
	serviceStopRole processRole = "service-stop"
)

type serviceStopMode string

const (
	stopTabServices serviceStopMode = "tabs"
	stopAllServices serviceStopMode = "all"
)

type processSpec struct {
	role  processRole
	args  []string
	dir   string
	env   []string
	input []byte
}

type processResult struct {
	code int
}

type childProcess interface {
	Done() <-chan processResult
	GroupAlive() bool
	SignalGroup(syscall.Signal) error
}

type processRunner interface {
	Start(processSpec) (childProcess, error)
}

type timer interface {
	Chan() <-chan time.Time
	Stop()
	Reset(time.Duration)
}

type clock interface {
	NewTimer(time.Duration) timer
}

type metadata interface {
	Reset() error
	RemoveReadiness() error
	SupervisionStarted() bool
	ArchiveDirectory(string) string
}

type coordinatorConfig struct {
	workloadArgs       []string
	workloadEnv        []string
	archiveEnv         []string
	archiveScript      []byte
	configurationError error
	spaceUser          string
	spaceHome          string
	spacePath          string
	playwrightPath     string
	archiveTimeout     time.Duration
	stopGrace          time.Duration
	serviceStopTimeout time.Duration
	serviceStopCommand string
	markerPollInterval time.Duration
	runner             processRunner
	clock              clock
	metadata           metadata
	stateChanged       func(lifecycleState)
	logger             *log.Logger
}

type coordinator struct {
	config coordinatorConfig
	state  lifecycleState

	workload    childProcess
	archive     childProcess
	serviceStop childProcess
	// Keep the process objects after their direct children exit so their
	// process groups can still be signaled if descendants remain.
	workloadExited    bool
	archiveExited     bool
	serviceStopExited bool
	serviceStopMode   serviceStopMode

	archiveRequested bool
	pendingArchive   bool
	archiveTimedOut  bool
	finalCode        int
	finalCodeSet     bool

	markerTimer      timer
	archiveTimer     timer
	archiveKillTimer timer
	processKillTimer timer
	serviceStopTimer timer
}

type systemRunner struct{}

func (systemRunner) Start(spec processSpec) (childProcess, error) {
	if len(spec.args) == 0 {
		return nil, errors.New("process command is empty")
	}

	command := exec.Command(spec.args[0], spec.args[1:]...)
	command.Dir = spec.dir
	command.Env = spec.env
	if spec.input == nil {
		command.Stdin = os.Stdin
	} else {
		command.Stdin = bytes.NewReader(spec.input)
	}
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := command.Start(); err != nil {
		return nil, err
	}

	process := &systemProcess{
		pid:  command.Process.Pid,
		done: make(chan processResult, 1),
	}
	go func() {
		process.done <- processResult{code: exitCode(command.Wait())}
		close(process.done)
	}()
	return process, nil
}

type systemProcess struct {
	pid  int
	done chan processResult
}

func (process *systemProcess) Done() <-chan processResult {
	return process.done
}

func (process *systemProcess) GroupAlive() bool {
	err := syscall.Kill(-process.pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func (process *systemProcess) SignalGroup(signal syscall.Signal) error {
	err := syscall.Kill(-process.pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

type systemClock struct{}

func (systemClock) NewTimer(duration time.Duration) timer {
	return &systemTimer{timer: time.NewTimer(duration)}
}

type systemTimer struct {
	timer *time.Timer
}

func (timer *systemTimer) Chan() <-chan time.Time {
	return timer.timer.C
}

func (timer *systemTimer) Stop() {
	timer.timer.Stop()
}

func (timer *systemTimer) Reset(duration time.Duration) {
	timer.timer.Reset(duration)
}

type fileMetadata struct {
	lifecycleDirectory string
	readyFile          string
}

func (metadata fileMetadata) Reset() error {
	if err := os.MkdirAll(metadata.lifecycleDirectory, 0o755); err != nil {
		return fmt.Errorf("create lifecycle metadata directory: %w", err)
	}
	if err := os.Chmod(metadata.lifecycleDirectory, 0o755); err != nil {
		return fmt.Errorf("set lifecycle metadata directory permissions: %w", err)
	}
	for _, path := range []string{
		filepath.Join(metadata.lifecycleDirectory, archiveDirectoryFilename),
		filepath.Join(metadata.lifecycleDirectory, workloadStartedFilename),
	} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove stale lifecycle metadata %q: %w", path, err)
		}
	}
	return nil
}

func (metadata fileMetadata) RemoveReadiness() error {
	err := os.Remove(metadata.readyFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (metadata fileMetadata) SupervisionStarted() bool {
	info, err := os.Stat(filepath.Join(metadata.lifecycleDirectory, workloadStartedFilename))
	return err == nil && info.Mode().IsRegular()
}

func (metadata fileMetadata) ArchiveDirectory(spaceHome string) string {
	contents, err := os.ReadFile(filepath.Join(metadata.lifecycleDirectory, archiveDirectoryFilename))
	if err != nil {
		return spaceHome
	}
	directory := string(contents)
	info, err := os.Stat(directory)
	if directory == "" || err != nil || !info.IsDir() {
		return spaceHome
	}
	return directory
}

func newCoordinator(config coordinatorConfig) *coordinator {
	return &coordinator{config: config}
}

func (coordinator *coordinator) Run(signals <-chan os.Signal) int {
	coordinator.setState(stateStarting)
	if err := coordinator.config.metadata.Reset(); err != nil {
		coordinator.config.logger.Printf("lifecycle: %v", err)
		coordinator.finish(1)
		return coordinator.finalCode
	}
	if err := coordinator.config.metadata.RemoveReadiness(); err != nil {
		coordinator.config.logger.Printf("lifecycle: remove readiness: %v", err)
		coordinator.finish(1)
		return coordinator.finalCode
	}
	if coordinator.config.configurationError != nil {
		coordinator.config.logger.Printf("lifecycle: %v", coordinator.config.configurationError)
		coordinator.finish(1)
		return coordinator.finalCode
	}

	workload, err := coordinator.config.runner.Start(processSpec{
		role: workloadRole,
		args: coordinator.config.workloadArgs,
		env:  coordinator.config.workloadEnv,
	})
	if err != nil {
		coordinator.config.logger.Printf("lifecycle: start workload: %v", err)
		coordinator.finish(1)
		return coordinator.finalCode
	}
	coordinator.workload = workload
	coordinator.archiveExited = true
	coordinator.serviceStopExited = true
	if coordinator.config.markerPollInterval > 0 {
		coordinator.markerTimer = coordinator.config.clock.NewTimer(coordinator.config.markerPollInterval)
	}

	for coordinator.state != stateDone {
		select {
		case receivedSignal := <-signals:
			coordinator.handleSignal(receivedSignal)
		case result := <-coordinator.workloadDone():
			coordinator.handleWorkloadExit(result)
		case result := <-coordinator.archiveDone():
			coordinator.handleArchiveExit(result)
		case result := <-coordinator.serviceStopDone():
			coordinator.handleServiceStopExit(result)
		case <-coordinator.markerTimerChannel():
			coordinator.handleMarkerTimer()
		case <-coordinator.archiveTimerChannel():
			coordinator.handleArchiveTimeout()
		case <-coordinator.archiveKillTimerChannel():
			coordinator.handleArchiveKillTimer()
		case <-coordinator.processKillTimerChannel():
			coordinator.handleProcessKillTimer()
		case <-coordinator.serviceStopTimerChannel():
			coordinator.handleServiceStopTimeout()
		}
	}
	return coordinator.finalCode
}

func (coordinator *coordinator) handleSignal(received os.Signal) {
	signal, ok := received.(syscall.Signal)
	if !ok {
		return
	}

	switch signal {
	case syscall.SIGUSR1:
		coordinator.requestArchive()
	case syscall.SIGTERM, syscall.SIGINT:
		coordinator.forceStop(signal)
	}
}

func (coordinator *coordinator) requestArchive() {
	if coordinator.archiveRequested || (coordinator.state != stateStarting && coordinator.state != stateRunning) {
		return
	}
	coordinator.archiveRequested = true
	if err := coordinator.config.metadata.RemoveReadiness(); err != nil {
		coordinator.config.logger.Printf("lifecycle: remove readiness: %v", err)
	}

	if coordinator.state == stateStarting && coordinator.config.metadata.SupervisionStarted() {
		coordinator.setState(stateRunning)
		coordinator.stopTimer(&coordinator.markerTimer)
	}

	if len(coordinator.config.archiveScript) == 0 {
		coordinator.beginFinalStop(1, syscall.SIGTERM)
		return
	}
	if coordinator.state == stateStarting {
		coordinator.pendingArchive = true
		coordinator.setState(stateStopping)
		coordinator.stopTimer(&coordinator.markerTimer)
		coordinator.signalProcess(coordinator.workload, syscall.SIGTERM, "bootstrap workload")
		coordinator.startProcessKillTimer()
		return
	}
	coordinator.startServiceStop(stopTabServices)
}

func (coordinator *coordinator) forceStop(signal syscall.Signal) {
	if coordinator.state == stateDone {
		return
	}
	coordinator.pendingArchive = false
	coordinator.stopTimer(&coordinator.markerTimer)
	coordinator.stopTimer(&coordinator.archiveTimer)
	coordinator.stopTimer(&coordinator.archiveKillTimer)
	coordinator.beginFinalStop(128+int(signal), signal)
}

func (coordinator *coordinator) handleWorkloadExit(result processResult) {
	coordinator.workloadExited = true

	switch coordinator.state {
	case stateStarting, stateRunning:
		coordinator.finish(result.code)
	case stateArchiving:
		// An accepted archive request owns the final status. Keep the
		// coordinator alive until that attempt completes or times out.
	case stateStopping:
		if coordinator.pendingArchive {
			if coordinator.workload.GroupAlive() {
				if coordinator.config.markerPollInterval > 0 {
					coordinator.markerTimer = coordinator.config.clock.NewTimer(coordinator.config.markerPollInterval)
				}
				return
			}
			coordinator.startPendingArchive()
			return
		}
		coordinator.finishIfStopped()
	}
}

func (coordinator *coordinator) handleServiceStopExit(result processResult) {
	coordinator.serviceStopExited = true
	coordinator.stopTimer(&coordinator.serviceStopTimer)
	if result.code != 0 {
		coordinator.config.logger.Printf("lifecycle: service stop exited with status %d", result.code)
	}

	if coordinator.finalCodeSet {
		coordinator.finishIfStopped()
		return
	}
	if coordinator.archiveRequested && coordinator.state == stateStopping {
		coordinator.startArchive()
	}
}

func (coordinator *coordinator) handleArchiveExit(result processResult) {
	coordinator.archiveExited = true
	coordinator.stopTimer(&coordinator.archiveTimer)

	if coordinator.state == stateArchiving {
		if coordinator.archiveTimedOut {
			if coordinator.archive.GroupAlive() && coordinator.archiveKillTimer != nil {
				return
			}
			coordinator.stopTimer(&coordinator.archiveKillTimer)
			coordinator.beginFinalStop(124, syscall.SIGTERM)
			return
		}
		coordinator.stopTimer(&coordinator.archiveKillTimer)
		coordinator.beginFinalStop(result.code, syscall.SIGTERM)
		return
	}
	coordinator.stopTimer(&coordinator.archiveKillTimer)
	if coordinator.state == stateStopping {
		coordinator.finishIfStopped()
	}
}

func (coordinator *coordinator) handleMarkerTimer() {
	if coordinator.pendingArchive {
		if coordinator.workload.GroupAlive() {
			coordinator.markerTimer.Reset(coordinator.config.markerPollInterval)
			return
		}
		coordinator.startPendingArchive()
		return
	}
	if coordinator.state != stateStarting {
		coordinator.stopTimer(&coordinator.markerTimer)
		return
	}
	if coordinator.config.metadata.SupervisionStarted() {
		coordinator.setState(stateRunning)
		coordinator.stopTimer(&coordinator.markerTimer)
		return
	}
	coordinator.markerTimer.Reset(coordinator.config.markerPollInterval)
}

func (coordinator *coordinator) startPendingArchive() {
	coordinator.pendingArchive = false
	coordinator.stopTimer(&coordinator.markerTimer)
	coordinator.stopTimer(&coordinator.processKillTimer)
	coordinator.startArchive()
}

func (coordinator *coordinator) startArchive() {
	coordinator.setState(stateArchiving)
	workingDirectory := coordinator.config.metadata.ArchiveDirectory(coordinator.config.spaceHome)
	archive, err := coordinator.config.runner.Start(processSpec{
		role:  archiveRole,
		dir:   workingDirectory,
		env:   coordinator.config.archiveEnv,
		input: coordinator.config.archiveScript,
		args: []string{
			"gosu",
			coordinator.config.spaceUser,
			"env",
			"HOME=" + coordinator.config.spaceHome,
			"USER=" + coordinator.config.spaceUser,
			"LOGNAME=" + coordinator.config.spaceUser,
			"PATH=" + coordinator.config.spacePath,
			"PLAYWRIGHT_BROWSERS_PATH=" + coordinator.config.playwrightPath,
			"bash",
			"--login",
			"-s",
		},
	})
	if err != nil {
		coordinator.config.logger.Printf("lifecycle: start archive: %v", err)
		coordinator.beginFinalStop(1, syscall.SIGTERM)
		return
	}
	coordinator.archive = archive
	coordinator.archiveExited = false
	coordinator.archiveTimer = coordinator.config.clock.NewTimer(coordinator.config.archiveTimeout)
}

func (coordinator *coordinator) startServiceStop(mode serviceStopMode) {
	coordinator.setState(stateStopping)
	if coordinator.config.serviceStopCommand == "" {
		coordinator.serviceStopExited = true
		if coordinator.finalCodeSet {
			coordinator.finishIfStopped()
		} else {
			coordinator.startArchive()
		}
		return
	}

	serviceStop, err := coordinator.config.runner.Start(processSpec{
		role: serviceStopRole,
		args: []string{coordinator.config.serviceStopCommand, string(mode)},
		env:  coordinator.config.archiveEnv,
	})
	if err != nil {
		coordinator.config.logger.Printf("lifecycle: start service stop: %v", err)
		coordinator.serviceStopExited = true
		if coordinator.finalCodeSet {
			coordinator.finishIfStopped()
		} else {
			coordinator.startArchive()
		}
		return
	}
	coordinator.serviceStop = serviceStop
	coordinator.serviceStopExited = false
	coordinator.serviceStopMode = mode
	coordinator.serviceStopTimer = coordinator.config.clock.NewTimer(coordinator.config.serviceStopTimeout)
}

func (coordinator *coordinator) handleServiceStopTimeout() {
	coordinator.stopTimer(&coordinator.serviceStopTimer)
	if coordinator.serviceStopExited {
		return
	}
	coordinator.config.logger.Printf("lifecycle: service stop timed out")
	coordinator.signalProcess(coordinator.serviceStop, syscall.SIGKILL, "service stop")
	coordinator.serviceStopExited = true
	if coordinator.finalCodeSet {
		coordinator.finishIfStopped()
	} else {
		coordinator.startArchive()
	}
}

func (coordinator *coordinator) handleArchiveTimeout() {
	if coordinator.state != stateArchiving || coordinator.archiveExited {
		coordinator.stopTimer(&coordinator.archiveTimer)
		return
	}
	select {
	case result := <-coordinator.archive.Done():
		coordinator.handleArchiveExit(result)
		return
	default:
	}

	coordinator.archiveTimedOut = true
	coordinator.stopTimer(&coordinator.archiveTimer)
	coordinator.signalProcess(coordinator.archive, syscall.SIGTERM, "archive")
	coordinator.archiveKillTimer = coordinator.config.clock.NewTimer(coordinator.config.stopGrace)
}

func (coordinator *coordinator) handleArchiveKillTimer() {
	coordinator.stopTimer(&coordinator.archiveKillTimer)
	if coordinator.state == stateArchiving && coordinator.archive.GroupAlive() {
		coordinator.signalProcess(coordinator.archive, syscall.SIGKILL, "archive")
	}
	if coordinator.state == stateArchiving && coordinator.archiveExited {
		coordinator.beginFinalStop(124, syscall.SIGTERM)
	}
}

func (coordinator *coordinator) handleProcessKillTimer() {
	coordinator.stopTimer(&coordinator.processKillTimer)
	coordinator.signalProcess(coordinator.archive, syscall.SIGKILL, "archive")
	coordinator.signalProcess(coordinator.workload, syscall.SIGKILL, "workload")
	if coordinator.pendingArchive {
		coordinator.pendingArchive = false
		coordinator.startArchive()
	}
}

func (coordinator *coordinator) beginFinalStop(code int, signal syscall.Signal) {
	coordinator.pendingArchive = false
	coordinator.finalCode = code
	coordinator.finalCodeSet = true
	coordinator.setState(stateStopping)
	coordinator.signalProcess(coordinator.archive, signal, "archive")
	coordinator.signalProcess(coordinator.workload, signal, "workload")
	if !coordinator.serviceStopExited && coordinator.serviceStopMode != stopAllServices {
		coordinator.signalProcess(coordinator.serviceStop, syscall.SIGKILL, "service stop")
		coordinator.stopTimer(&coordinator.serviceStopTimer)
		coordinator.serviceStopExited = true
	}
	if coordinator.serviceStopExited {
		coordinator.startServiceStop(stopAllServices)
	}
	if !coordinator.archiveExited || !coordinator.workloadExited {
		coordinator.startProcessKillTimer()
	}
	coordinator.finishIfStopped()
}

func (coordinator *coordinator) finishIfStopped() {
	if coordinator.state == stateStopping && coordinator.finalCodeSet &&
		coordinator.workloadExited && coordinator.archiveExited && coordinator.serviceStopExited {
		coordinator.finish(coordinator.finalCode)
	}
}

func (coordinator *coordinator) finish(code int) {
	coordinator.stopTimer(&coordinator.markerTimer)
	coordinator.stopTimer(&coordinator.archiveTimer)
	coordinator.stopTimer(&coordinator.archiveKillTimer)
	coordinator.stopTimer(&coordinator.processKillTimer)
	coordinator.stopTimer(&coordinator.serviceStopTimer)
	coordinator.finalCode = code
	coordinator.finalCodeSet = true
	coordinator.setState(stateDone)
}

func (coordinator *coordinator) startProcessKillTimer() {
	coordinator.stopTimer(&coordinator.processKillTimer)
	coordinator.processKillTimer = coordinator.config.clock.NewTimer(coordinator.config.stopGrace)
}

func (coordinator *coordinator) signalProcess(process childProcess, signal syscall.Signal, description string) {
	if process == nil || !process.GroupAlive() {
		return
	}
	if err := process.SignalGroup(signal); err != nil {
		coordinator.config.logger.Printf("lifecycle: signal %s with %s: %v", description, signal, err)
	}
}

func (coordinator *coordinator) setState(state lifecycleState) {
	if coordinator.state == state {
		return
	}
	coordinator.state = state
	if coordinator.config.stateChanged != nil {
		coordinator.config.stateChanged(state)
	}
}

func (coordinator *coordinator) stopTimer(target *timer) {
	if *target == nil {
		return
	}
	(*target).Stop()
	*target = nil
}

func (coordinator *coordinator) workloadDone() <-chan processResult {
	if coordinator.workload == nil || coordinator.workloadExited {
		return nil
	}
	return coordinator.workload.Done()
}

func (coordinator *coordinator) archiveDone() <-chan processResult {
	if coordinator.archive == nil || coordinator.archiveExited {
		return nil
	}
	return coordinator.archive.Done()
}

func (coordinator *coordinator) serviceStopDone() <-chan processResult {
	if coordinator.serviceStop == nil || coordinator.serviceStopExited {
		return nil
	}
	return coordinator.serviceStop.Done()
}

func (coordinator *coordinator) markerTimerChannel() <-chan time.Time {
	if coordinator.markerTimer == nil {
		return nil
	}
	return coordinator.markerTimer.Chan()
}

func (coordinator *coordinator) archiveTimerChannel() <-chan time.Time {
	if coordinator.archiveTimer == nil {
		return nil
	}
	return coordinator.archiveTimer.Chan()
}

func (coordinator *coordinator) archiveKillTimerChannel() <-chan time.Time {
	if coordinator.archiveKillTimer == nil {
		return nil
	}
	return coordinator.archiveKillTimer.Chan()
}

func (coordinator *coordinator) processKillTimerChannel() <-chan time.Time {
	if coordinator.processKillTimer == nil {
		return nil
	}
	return coordinator.processKillTimer.Chan()
}

func (coordinator *coordinator) serviceStopTimerChannel() <-chan time.Time {
	if coordinator.serviceStopTimer == nil {
		return nil
	}
	return coordinator.serviceStopTimer.Chan()
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) {
		return 1
	}
	status, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok {
		return 1
	}
	if status.Signaled() {
		return 128 + int(status.Signal())
	}
	return status.ExitStatus()
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func decodeScriptEnvironment(name string) ([]byte, error) {
	encoded := os.Getenv(name)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return decoded, nil
}

func filteredEnvironment(environment []string, exclude func(string) bool) []string {
	filtered := make([]string, 0, len(environment))
	for _, variable := range environment {
		name, _, _ := strings.Cut(variable, "=")
		if !exclude(name) {
			filtered = append(filtered, variable)
		}
	}
	return filtered
}

func excludedWorkloadEnvironment(name string) bool {
	return name == "CLOR_ARCHIVE_SCRIPT_BASE64"
}

func excludedArchiveEnvironment(name string) bool {
	if excludedWorkloadEnvironment(name) || name == "CLOR_SETUP_SCRIPT_BASE64" || name == "CLOR_API_KEY" {
		return true
	}
	return strings.HasPrefix(name, "CLOR_TAB_") &&
		(strings.HasSuffix(name, "_WEBTUI_SECRET") || strings.HasSuffix(name, "_DIRECT_TOKEN"))
}

func configFromEnvironment(workloadArgs []string) coordinatorConfig {
	spaceUser := environmentOrDefault("CLOR_SPACE_USER", defaultSpaceUser)
	spaceHome := environmentOrDefault("CLOR_SPACE_HOME", defaultSpaceHome)
	spacePath := fmt.Sprintf("%s/.local/bin:%s/go/bin:%s/.npm-global/bin", spaceHome, spaceHome, spaceHome)
	if prefix := os.Getenv("CLOR_SPACE_PATH_PREFIX"); prefix != "" {
		spacePath += ":" + prefix
	}
	spacePath += ":" + os.Getenv("PATH")
	_, setupScriptError := decodeScriptEnvironment("CLOR_SETUP_SCRIPT_BASE64")
	archiveScript, archiveScriptError := decodeScriptEnvironment("CLOR_ARCHIVE_SCRIPT_BASE64")
	environment := os.Environ()

	return coordinatorConfig{
		workloadArgs:       workloadArgs,
		workloadEnv:        filteredEnvironment(environment, excludedWorkloadEnvironment),
		archiveEnv:         filteredEnvironment(environment, excludedArchiveEnvironment),
		archiveScript:      archiveScript,
		configurationError: errors.Join(setupScriptError, archiveScriptError),
		spaceUser:          spaceUser,
		spaceHome:          spaceHome,
		spacePath:          spacePath,
		playwrightPath:     filepath.Join(spaceHome, ".cache/ms-playwright"),
		archiveTimeout:     defaultArchiveTimeout,
		stopGrace:          defaultStopGrace,
		serviceStopTimeout: defaultServiceStopTimeout,
		serviceStopCommand: defaultServiceStopCommand,
		markerPollInterval: defaultMarkerPoll,
		runner:             systemRunner{},
		clock:              systemClock{},
		metadata: fileMetadata{
			lifecycleDirectory: defaultLifecycleDir,
			readyFile:          defaultReadyFile,
		},
		logger: log.New(os.Stderr, "", 0),
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "lifecycle: workload command is required")
		os.Exit(1)
	}

	config := configFromEnvironment(os.Args[1:])
	signals := make(chan os.Signal, 16)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGINT)
	code := newCoordinator(config).Run(signals)
	signal.Stop(signals)
	os.Exit(code)
}
