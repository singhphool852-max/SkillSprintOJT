package judge

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"os"
)

// ──────────────────────────────────────────────
// Executor — pluggable code execution interface.
//
// Implementations:
//   - LocalExecutor:  runs in Docker containers (production)
//   - CloudExecutor:  sends to remote worker fleet (future)
//
// To swap: change NewDefaultExecutor() in service.go.
// ──────────────────────────────────────────────

// Executor defines how code is run against testcases.
type Executor interface {
	Run(code, language, input string, timeLimitMs int) (ExecutionResult, error)
	SupportedLanguages() []LanguageInfo
}

// ExecutionResult holds the output from a single code execution.
type ExecutionResult struct {
	Output     string `json:"output"`
	ExitCode   int    `json:"exitCode"`
	TimedOut   bool   `json:"timedOut"`
	Error      string `json:"error"`      // empty means success
	ErrorType  string `json:"errorType"`  // compilation_error, runtime_error, time_limit_exceeded
	CompileOut string `json:"compileOut"` // compilation output (errors/warnings)
	DurationMs int64  `json:"durationMs"` // wall-clock execution time
}

// LanguageInfo describes a supported language.
type LanguageInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Template string `json:"template"`
}

// LocalExecutor runs code in isolated Docker containers with resource limits.
type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (e *LocalExecutor) SupportedLanguages() []LanguageInfo {
	return []LanguageInfo{
		{ID: "python", Name: "Python 3", Version: "3.11", Template: "import sys\n\ndef solve():\n    # Read input\n    line = input()\n    # Your solution here\n    print(line)\n\nsolve()\n"},
		{ID: "cpp", Name: "C++", Version: "C++17", Template: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    // Read input\n    string s;\n    getline(cin, s);\n    // Your solution here\n    cout << s << endl;\n    return 0;\n}\n"},
		{ID: "java", Name: "Java", Version: "21", Template: "import java.util.Scanner;\n\npublic class Main {\n    public static void main(String[] args) {\n        Scanner sc = new Scanner(System.in);\n        // Read input\n        String s = sc.nextLine();\n        // Your solution here\n        System.out.println(s);\n    }\n}\n"},
		{ID: "javascript", Name: "JavaScript", Version: "Node.js 20", Template: "const readline = require('readline');\nconst rl = readline.createInterface({ input: process.stdin });\nconst lines = [];\nrl.on('line', (line) => lines.push(line));\nrl.on('close', () => {\n    // Your solution here\n    console.log(lines[0]);\n});\n"},
		{ID: "go", Name: "Go", Version: "1.21", Template: "package main\n\nimport (\n\t\"bufio\"\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tscanner := bufio.NewScanner(os.Stdin)\n\tscanner.Scan()\n\t// Your solution here\n\tfmt.Println(scanner.Text())\n}\n"},
	}
}

// Run executes code in an isolated Docker container with security constraints.
// Each execution:
// - Creates isolated temp directory
// - Writes code to file
// - Runs in Docker with: no network, memory limit, CPU limit, read-only code mount
// - 10 second timeout
// - Captures stdout for comparison
// - Cleans up temp directory
func (e *LocalExecutor) Run(code, language, input string, timeLimitMs int) (ExecutionResult, error) {
	log.Printf("[EXECUTOR] executing language: %s", language)
	
	if timeLimitMs <= 0 {
		timeLimitMs = 2000
	}

	// Create isolated temp directory for this execution
	tmpDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return ExecutionResult{Error: "failed to create temp dir", ErrorType: "internal_error"}, err
	}
	defer os.RemoveAll(tmpDir)

	var filename, image string
	var runCmd []string

	// Language to Docker image mapping
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "python", "python3":
		filename = "solution.py"
		image = "python:3.11-alpine"
		runCmd = []string{"python3", "/code/solution.py"}
		
	case "go", "golang":
		filename = "main.go"
		image = "golang:1.21-alpine"
		runCmd = []string{"go", "run", "/code/main.go"}
		
	case "cpp", "c++", "c":
		filename = "solution.cpp"
		image = "gcc:13"
		runCmd = []string{"sh", "-c", "g++ -O2 -o /tmp/solution /code/solution.cpp && /tmp/solution"}
		
	case "java":
		filename = "Main.java"
		image = "openjdk:21-slim"
		runCmd = []string{"sh", "-c", "javac /code/Main.java && java -cp /code Main"}
		
	case "javascript", "js", "node":
		filename = "solution.js"
		image = "node:20-alpine"
		runCmd = []string{"node", "/code/solution.js"}
		
	default:
		return ExecutionResult{
			Error:     fmt.Sprintf("unsupported language: %s", language),
			ErrorType: "unsupported_language",
		}, fmt.Errorf("unsupported language: %s", language)
	}

	// Write code to temp directory
	filePath := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return ExecutionResult{Error: "failed to write code file", ErrorType: "internal_error"}, err
	}

	// Create timeout context (10 seconds max)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build docker run command with security constraints
	args := []string{
		"run", "--rm",
		"--network", "none",           // No network access
		"--memory", "128m",            // 128MB RAM limit
		"--memory-swap", "128m",       // No swap
		"--cpus", "0.5",               // 50% CPU limit
		"--pids-limit", "50",          // Max 50 processes
		"-v", tmpDir + ":/code:ro",    // Mount code as read-only
		"-i",                          // Interactive (for stdin)
		image,
	}
	args = append(args, runCmd...)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = strings.NewReader(input)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Run(); err != nil {
		duration := time.Since(startTime).Milliseconds()
		
		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			return ExecutionResult{
				Output:     "TIME LIMIT EXCEEDED",
				TimedOut:   true,
				Error:      "time_limit_exceeded",
				ErrorType:  "time_limit_exceeded",
				DurationMs: 10000,
			}, nil
		}
		
		// Runtime error
		return ExecutionResult{
			Output:     strings.TrimSpace(stderr.String()),
			Error:      "runtime_error",
			ErrorType:  "runtime_error",
			DurationMs: duration,
		}, fmt.Errorf("runtime error: %s", strings.TrimSpace(stderr.String()))
	}
	
	duration := time.Since(startTime).Milliseconds()

	return ExecutionResult{
		Output:     strings.TrimSpace(stdout.String()),
		ExitCode:   0,
		DurationMs: duration,
	}, nil
}
