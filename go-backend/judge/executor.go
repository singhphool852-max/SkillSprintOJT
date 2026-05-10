package judge

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ──────────────────────────────────────────────
// Executor — pluggable code execution interface.
// The backend always programs against this interface.
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
	// SupportedLanguages returns languages this executor can handle.
	SupportedLanguages() []LanguageInfo
}

// ExecutionResult holds the output from a single code execution.
type ExecutionResult struct {
	Output      string `json:"output"`
	ExitCode    int    `json:"exitCode"`
	TimedOut    bool   `json:"timedOut"`
	Error       string `json:"error"`       // empty means success
	ErrorType   string `json:"errorType"`   // compilation_error, runtime_error, time_limit_exceeded
	CompileOut  string `json:"compileOut"`   // compilation output (errors/warnings)
	DurationMs  int64  `json:"durationMs"`  // wall-clock execution time
}

// LanguageInfo describes a supported language.
type LanguageInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Template string `json:"template"` // default starter code
}

// LocalExecutor runs code as a local subprocess with timeout enforcement.
// Supports Python 3, C++, Java, JavaScript (Node), and Go.
type LocalExecutor struct{}

// NewLocalExecutor returns a new subprocess-based executor.
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

// SupportedLanguages returns all languages this executor can handle.
func (e *LocalExecutor) SupportedLanguages() []LanguageInfo {
	return []LanguageInfo{
		{ID: "python", Name: "Python 3", Version: "3.x", Template: "import sys\n\ndef solve():\n    # Read input\n    line = input()\n    # Your solution here\n    print(line)\n\nsolve()\n"},
		{ID: "cpp", Name: "C++", Version: "C++17", Template: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    // Read input\n    string s;\n    getline(cin, s);\n    // Your solution here\n    cout << s << endl;\n    return 0;\n}\n"},
		{ID: "java", Name: "Java", Version: "17+", Template: "import java.util.Scanner;\n\npublic class Solution {\n    public static void main(String[] args) {\n        Scanner sc = new Scanner(System.in);\n        // Read input\n        String s = sc.nextLine();\n        // Your solution here\n        System.out.println(s);\n    }\n}\n"},
		{ID: "javascript", Name: "JavaScript", Version: "Node.js", Template: "const readline = require('readline');\nconst rl = readline.createInterface({ input: process.stdin });\nconst lines = [];\nrl.on('line', (line) => lines.push(line));\nrl.on('close', () => {\n    // Your solution here\n    console.log(lines[0]);\n});\n"},
		{ID: "go", Name: "Go", Version: "1.21+", Template: "package main\n\nimport (\n\t\"bufio\"\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tscanner := bufio.NewScanner(os.Stdin)\n\tscanner.Scan()\n\t// Your solution here\n\tfmt.Println(scanner.Text())\n}\n"},
	}
}

// Run executes the given code with the given input and time limit.
// It writes code to a temp file, compiles if needed, and runs with stdin piped.
func (e *LocalExecutor) Run(code, language, input string, timeLimitMs int) (ExecutionResult, error) {
	if timeLimitMs <= 0 {
		timeLimitMs = 2000
	}

	tmpDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return ExecutionResult{Error: "failed to create temp dir", ErrorType: "internal_error"}, err
	}
	defer os.RemoveAll(tmpDir)

	timeout := time.Duration(timeLimitMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	startTime := time.Now()

	switch strings.ToLower(language) {
	case "python", "python3", "py":
		srcPath := filepath.Join(tmpDir, "solution.py")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		cmd = exec.CommandContext(ctx, "python3", srcPath)

	case "cpp", "c++":
		srcPath := filepath.Join(tmpDir, "solution.cpp")
		binPath := filepath.Join(tmpDir, "solution")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		// Compile
		compileCtx, compileCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer compileCancel()
		compile := exec.CommandContext(compileCtx, "g++", "-o", binPath, srcPath, "-std=c++17")
		compileOut, compileErr := compile.CombinedOutput()
		if compileErr != nil {
			return ExecutionResult{
				Output:     string(compileOut),
				CompileOut: string(compileOut),
				ExitCode:   1,
				Error:      "compilation_error",
				ErrorType:  "compilation_error",
				DurationMs: time.Since(startTime).Milliseconds(),
			}, nil // Not an internal error — return compilation failure as result
		}
		cmd = exec.CommandContext(ctx, binPath)

	case "java":
		// Detect class name from code (e.g., "public class Main" → "Main")
		className := "Solution"
		re := regexp.MustCompile(`public\s+class\s+(\w+)`)
		if m := re.FindStringSubmatch(code); len(m) > 1 {
			className = m[1]
		}
		srcPath := filepath.Join(tmpDir, className+".java")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		// Compile
		compileCtx, compileCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer compileCancel()
		compile := exec.CommandContext(compileCtx, "javac", srcPath)
		compileOut, compileErr := compile.CombinedOutput()
		if compileErr != nil {
			return ExecutionResult{
				Output:     string(compileOut),
				CompileOut: string(compileOut),
				ExitCode:   1,
				Error:      "compilation_error",
				ErrorType:  "compilation_error",
				DurationMs: time.Since(startTime).Milliseconds(),
			}, nil
		}
		cmd = exec.CommandContext(ctx, "java", "-cp", tmpDir, className)

	case "javascript", "js", "node":
		srcPath := filepath.Join(tmpDir, "solution.js")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		cmd = exec.CommandContext(ctx, "node", srcPath)

	case "go", "golang":
		srcPath := filepath.Join(tmpDir, "main.go")
		binPath := filepath.Join(tmpDir, "solution_go")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{Error: "failed to write source", ErrorType: "internal_error"}, err
		}
		// Two-phase: compile with generous timeout, then run with user time limit
		compileCtx, compileCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer compileCancel()
		compile := exec.CommandContext(compileCtx, "go", "build", "-o", binPath, srcPath)
		compileOut, compileErr := compile.CombinedOutput()
		if compileErr != nil {
			return ExecutionResult{
				Output:     string(compileOut),
				CompileOut: string(compileOut),
				ExitCode:   1,
				Error:      "compilation_error",
				ErrorType:  "compilation_error",
				DurationMs: time.Since(startTime).Milliseconds(),
			}, nil
		}
		// Reset start time to exclude compilation from execution time
		startTime = time.Now()
		cmd = exec.CommandContext(ctx, binPath)

	default:
		return ExecutionResult{
			Error:     fmt.Sprintf("unsupported language: %s", language),
			ErrorType: "unsupported_language",
		}, fmt.Errorf("unsupported language: %s", language)
	}

	// Pipe stdin
	cmd.Stdin = strings.NewReader(input)

	// Capture stdout and stderr separately — critical for correct output comparison.
	// CombinedOutput mixes stderr warnings/info into stdout, causing wrong verdicts.
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()
	duration := time.Since(startTime).Milliseconds()

	// Check for timeout
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
	exitCode := 0
	if runErr != nil {
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

	// Success — return only stdout (stderr is ignored for output comparison)
	return ExecutionResult{
		Output:     stdout,
		ExitCode:   0,
		DurationMs: duration,
	}, nil
}
