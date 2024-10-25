package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"log"
	"bytes"
	"io"
)

// ExecutionLog stores detailed information about the execution
type ExecutionLog struct {
	ContainerID  string
	StartTime    time.Time
	EndTime      time.Time
	ExitCode     int
	StdOut       string
	StdErr       string
	DockerLogs   string
	Error        error
}

type Executor struct {
	volumePath  string
	logger      *log.Logger
}

func New() (*Executor, error) {
	// Create log file with append mode
    logPath := "./snow.log"
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Executor{
		volumePath: "/opt/libs",
		logger:     log.New(logFile, "", log.LstdFlags),
	}, nil
}


type ExecutionRequest struct {
	Language string
	Code     string
	Timeout  time.Duration
}

type ExecutionResult struct {
	Output string
	Error  error
	Logs   ExecutionLog
}

func (e *Executor) Execute(ctx context.Context, req ExecutionRequest) ExecutionResult {
	execLog := ExecutionLog{
		StartTime: time.Now(),
	}

	// Create temporary directory for execution
	tmpDir, err := os.MkdirTemp("", "execution-*")
	if err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(tmpDir)

	// Write code to file
	filename := filepath.Join(tmpDir, "code"+e.getFileExtension(req.Language))
	if err := os.WriteFile(filename, []byte(req.Code), 0644); err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to write code file: %w", err))
	}

	// Get docker configuration
	dockerImage := e.getDockerImage(req.Language)
	if dockerImage == "" {
		return e.handleError(execLog, fmt.Errorf("unsupported language: %s", req.Language))
	}

	runCommand := e.getRunCommand(req.Language)
	if runCommand == "" {
		return e.handleError(execLog, fmt.Errorf("unsupported language command: %s", req.Language))
	}

	// First, create the container
	createCmd := exec.Command("docker", "create",
		"--memory=512m",
		"--memory-swap=512m",
		"--cpus=1",
		"--ulimit", "nproc=1024:1024",
		"--ulimit", "nofile=1024:1024",
		"--pids-limit=100",
		"--network=none",
		"-v", e.volumePath+":/opt/libs:ro",
		"-v", tmpDir+":/code",
		"-w", "/code",
		dockerImage,
		"sh", "-c", fmt.Sprintf("%s code%s", runCommand, e.getFileExtension(req.Language)))

	containerIDBytes, err := createCmd.Output()
	if err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to create container: %w", err))
	}

	containerID := string(bytes.TrimSpace(containerIDBytes))
	execLog.ContainerID = containerID
	e.logger.Printf("Created container: %s", containerID)

	// Start the container
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	startCmd := exec.CommandContext(ctx, "docker", "start", "-a", containerID)
	
	// Create pipes for stdout and stderr
	stdoutPipe, err := startCmd.StdoutPipe()
	if err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to create stdout pipe: %w", err))
	}
	stderrPipe, err := startCmd.StderrPipe()
	if err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to create stderr pipe: %w", err))
	}

	// Start the command
	if err := startCmd.Start(); err != nil {
		return e.handleError(execLog, fmt.Errorf("failed to start container: %w", err))
	}

	// Capture stdout and stderr
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	
	go io.Copy(stdout, stdoutPipe)
	go io.Copy(stderr, stderrPipe)

	// Wait for completion
	err = startCmd.Wait()
	execLog.EndTime = time.Now()
	
	// Get container logs
	logsCmd := exec.Command("docker", "logs", containerID)
	dockerLogs, _ := logsCmd.CombinedOutput()
	execLog.DockerLogs = string(dockerLogs)

	// Get exit code
	if exitErr, ok := err.(*exec.ExitError); ok {
		execLog.ExitCode = exitErr.ExitCode()
	}

	// Cleanup container
	cleanupCmd := exec.Command("docker", "rm", "-f", containerID)
	_ = cleanupCmd.Run()

	// Store outputs
	execLog.StdOut = stdout.String()
	execLog.StdErr = stderr.String()
	execLog.Error = err

	// Log execution details
	e.logger.Printf("Execution completed for container %s:\n"+
		"Duration: %v\n"+
		"Exit Code: %d\n"+
		"Error: %v\n"+
		"Docker Logs: %s\n"+
		"StdOut: %s\n"+
		"StdErr: %s\n",
		containerID,
		execLog.EndTime.Sub(execLog.StartTime),
		execLog.ExitCode,
		execLog.Error,
		execLog.DockerLogs,
		execLog.StdOut,
		execLog.StdErr)

	return ExecutionResult{
		Output: execLog.StdOut,
		Error:  execLog.Error,
		Logs:   execLog,
	}
}

func (e *Executor) handleError(log ExecutionLog, err error) ExecutionResult {
	log.EndTime = time.Now()
	log.Error = err
	
	e.logger.Printf("Execution error: %v\n"+
		"Container ID: %s\n"+
		"Duration: %v\n",
		err,
		log.ContainerID,
		log.EndTime.Sub(log.StartTime))
	
	return ExecutionResult{
		Error: err,
		Logs:  log,
	}
}



func (e *Executor) getFileExtension(lang string) string {
	extensions := map[string]string{
		"python": ".py",
		"nodejs": ".js",
		"golang": ".go",
		"cpp":    ".cpp",
	}
	ext, ok := extensions[lang]
	if !ok {
		return ""
	}
	return ext
}

func (e *Executor) getDockerImage(lang string) string {
	images := map[string]string{
		"python": "python:3.9-slim",
		"nodejs": "node:16-slim",
		"golang": "golang:1.18-alpine",
		"cpp":    "gcc:latest",
	}
	image, ok := images[lang]
	if !ok {
		return ""
	}
	return image
}

func (e *Executor) getRunCommand(lang string) string {
	commands := map[string]string{
		"python": "python",
		"nodejs": "node",
		"golang": "go run",
		"cpp":    "g++ -o code.out code.cpp && ./code.out",
	}
	cmd, ok := commands[lang]
	if !ok {
		return ""
	}
	return cmd
}
