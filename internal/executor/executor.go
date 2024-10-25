package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Executor struct {
	volumePath string
}

func New() *Executor {
	return &Executor{ 
		volumePath: "/opt/libs",           
	}
}

type ExecutionRequest struct {
	Language string
	Code     string
	Timeout  time.Duration
}

type ExecutionResult struct {
	Output string
	Error  error
}

func (e *Executor) Execute(ctx context.Context, req ExecutionRequest) ExecutionResult {

	// Create temporary directory for execution
	tmpDir, err := os.MkdirTemp("", "execution-*")
	if err != nil {
		return ExecutionResult{Error: fmt.Errorf("failed to create temp dir: %w", err)}
	}
	defer os.RemoveAll(tmpDir)

	// Prepare code with language-specific requirements)

	// Write code to file
	filename := filepath.Join(tmpDir, "code"+e.getFileExtension(req.Language))
	if err := os.WriteFile(filename, []byte(req.Code), 0644); err != nil {
		return ExecutionResult{Error: fmt.Errorf("failed to write code file: %w", err)}
	}

	// Get docker configuration
	dockerImage := e.getDockerImage(req.Language)
	if dockerImage == "" {
		return ExecutionResult{Error: fmt.Errorf("unsupported language: %s", req.Language)}
	}

	runCommand := e.getRunCommand(req.Language)
	if runCommand == "" {
		return ExecutionResult{Error: fmt.Errorf("unsupported language command: %s", req.Language)}
	}

	// Prepare Docker command with adjusted resource limits
	dockerCmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",
		"--memory=512m",
		"--memory-swap=512m",
		"--cpus=1",
		"--ulimit", "nproc=1024:1024",    // Increase process limit
		"--ulimit", "nofile=1024:1024",   // Increase file descriptor limit
		"--pids-limit=100",               // Increased from 50
		"--network=none",
		"-v", e.volumePath+":/opt/libs:ro",
		"-v", tmpDir+":/code:ro",
		"-w", "/code",
		dockerImage,
		"sh", "-c", fmt.Sprintf("%s code%s", runCommand, e.getFileExtension(req.Language)))

	// Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	output, err := dockerCmd.CombinedOutput()

	if err != nil {
		return ExecutionResult{
			Output: string(output),
			Error:  fmt.Errorf("execution failed: %w", err),
		}
	}

	return ExecutionResult{
		Output: string(output),
		Error:  nil,
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
