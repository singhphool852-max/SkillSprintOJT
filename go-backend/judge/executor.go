package judge

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ──────────────────────────────────────────────
// Executor — pluggable code execution interface.
//
// Implementations:
//   - LocalExecutor:  runs on same machine (dev/small setups)
//   - DockerExecutor: runs inside ephemeral containers (future)
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

// LocalExecutor runs code as a local subprocess with timeout enforcement.
type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (e *LocalExecutor) SupportedLanguages() []LanguageInfo {
	return []LanguageInfo{
		{ID: "python", Name: "Python 3", Version: "3.x", Template: "import sys\n\ndef solve():\n    # Read input\n    line = input()\n    # Your solution here\n    print(line)\n\nsolve()\n"},
		{ID: "cpp", Name: "C++", Version: "C++17", Template: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    // Read input\n    string s;\n    getline(cin, s);\n    // Your solution here\n    cout << s << endl;\n    return 0;\n}\n"},
		{ID: "java", Name: "Java", Version: "17+", Template: "import java.util.Scanner;\n\npublic class Solution {\n    public static void main(String[] args) {\n        Scanner sc = new Scanner(System.in);\n        // Read input\n        String s = sc.nextLine();\n        // Your solution here\n        System.out.println(s);\n    }\n}\n"},
		{ID: "javascript", Name: "JavaScript", Version: "Node.js", Template: "const readline = require('readline');\nconst rl = readline.createInterface({ input: process.stdin });\nconst lines = [];\nrl.on('line', (line) => lines.push(line));\nrl.on('close', () => {\n    // Your solution here\n    console.log(lines[0]);\n});\n"},
		{ID: "go", Name: "Go", Version: "1.21+", Template: "package main\n\nimport (\n\t\"bufio\"\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tscanner := bufio.NewScanner(os.Stdin)\n\tscanner.Scan()\n\t// Your solution here\n\tfmt.Println(scanner.Text())\n}\n"},
	}
}

// normalizeInput strips \r from input so programs always see clean \n line endings.
func normalizeInput(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// normalizeOutput strips \r from output so comparison is consistent.
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// compile runs a compilation command with a generous timeout.
// Returns (compileOutput, error). If error is non-nil, compilation failed.
func compile(args []string, timeoutSec int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Run executes the given code with the given input and time limit.
// Each call uses its own isolated temp directory for concurrency safety.
// The execution timeout only covers the run phase, NOT compilation.
func (e *LocalExecutor) Run(code, language, input string, timeLimitMs int) (ExecutionResult, error) {
	log.Printf("[EXECUTOR] executing language: %s", language)
	
	if timeLimitMs <= 0 {
		timeLimitMs = 2000
	}

	// Each execution gets its own temp dir — safe for concurrent goroutines
	tmpDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return ExecutionResult{Error: "failed to create temp dir", ErrorType: "internal_error"}, err
	}
	defer os.RemoveAll(tmpDir)

	// Normalize input: strip \r so programs always see clean \n
	input = normalizeInput(input)

	// Determine what to run based on language.
	// For compiled languages: compile first (outside the execution timeout),
	// then run the binary within the timeout.
	var runArgs []string // command + args for the run phase
	var runDir string    // working directory for the run phase

	switch strings.ToLower(language) {
	case "python", "python3", "py":
		srcPath := filepath.Join(tmpDir, "solution.py")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		runArgs = []string{"python3", srcPath}

	case "cpp", "c++":
		srcPath := filepath.Join(tmpDir, "solution.cpp")
		binPath := filepath.Join(tmpDir, "solution")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		compOut, compErr := compile([]string{"g++", "-O2", "-o", binPath, srcPath, "-std=c++17"}, 15)
		if compErr != nil {
			return ExecutionResult{
				Output:     compOut,
				CompileOut: compOut,
				ExitCode:   1,
				Error:      "compilation_error",
				ErrorType:  "compilation_error",
			}, nil
		}
		runArgs = []string{binPath}

	case "java":
		// Detect class name from code
		className := "Solution"
		re := regexp.MustCompile(`public\s+class\s+(\w+)`)
		if m := re.FindStringSubmatch(code); len(m) > 1 {
			className = m[1]
		}
		srcPath := filepath.Join(tmpDir, className+".java")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		compOut, compErr := compile([]string{"javac", "-d", tmpDir, srcPath}, 15)
		if compErr != nil {
			return ExecutionResult{
				Output:     compOut,
				CompileOut: compOut,
				ExitCode:   1,
				Error:      "compilation_error",
				ErrorType:  "compilation_error",
			}, nil
		}
		runArgs = []string{"java", "-cp", tmpDir, className}

	case "javascript", "js", "node":
		srcPath := filepath.Join(tmpDir, "solution.js")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		runArgs = []string{"node", srcPath}

	case "go", "golang":
		srcPath := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		runArgs = []string{"go", "run", srcPath}
		runDir = tmpDir
		timeLimitMs += 10000

	default:
		return ExecutionResult{
			Error:     fmt.Sprintf("unsupported language: %s", language),
			ErrorType: "unsupported_language",
		}, fmt.Errorf("unsupported language: %s", language)
	}

	// ── Run phase: create a FRESH timeout context that starts NOW ──
	// This ensures compilation time is NOT counted against the user's time limit.
	timeout := time.Duration(timeLimitMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, runArgs[0], runArgs[1:]...)
	if runDir != "" {
		cmd.Dir = runDir
	}

	// Pipe test input to stdin
	cmd.Stdin = strings.NewReader(input)

	// Capture stdout and stderr SEPARATELY — only stdout is used for comparison
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	startTime := time.Now()
	runErr := cmd.Run()
	duration := time.Since(startTime).Milliseconds()

	stdout := normalizeOutput(stdoutBuf.String())
	stderr := stderrBuf.String()
	
	log.Printf("[EXECUTOR] lang=%s duration=%dms stderr=%q", language, duration, stderr)

	// Check for timeout FIRST
	if ctx.Err() == context.DeadlineExceeded {
		return ExecutionResult{
			Output:     stdout,
			CompileOut: stderr,
			ExitCode:   -1,
			TimedOut:   true,
			Error:      "time_limit_exceeded",
			ErrorType:  "time_limit_exceeded",
			DurationMs: duration,
		}, nil
	}

	// Check for runtime error
	if runErr != nil {
		exitCode := 0
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		// Include stderr in output for runtime errors so user sees the error message
		errOutput := stdout
		if stderr != "" {
			if errOutput != "" {
				errOutput += "\n"
			}
			errOutput += stderr
		}
		return ExecutionResult{
			Output:     errOutput,
			CompileOut: stderr,
			ExitCode:   exitCode,
			Error:      "runtime_error",
			ErrorType:  "runtime_error",
			DurationMs: duration,
		}, nil
	}

	// Success — return only stdout (stderr is NOT mixed into comparison output)
	return ExecutionResult{
		Output:     stdout,
		ExitCode:   0,
		DurationMs: duration,
	}, nil
}
