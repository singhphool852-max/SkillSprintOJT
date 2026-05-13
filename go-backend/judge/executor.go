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

type Executor interface {
	Run(code, language, input string, timeLimitMs int) (ExecutionResult, error)
	SupportedLanguages() []LanguageInfo
}

type ExecutionResult struct {
	Output     string `json:"output"`
	ExitCode   int    `json:"exitCode"`
	TimedOut   bool   `json:"timedOut"`
	Error      string `json:"error"`
	ErrorType  string `json:"errorType"`
	CompileOut string `json:"compileOut"`
	DurationMs int64  `json:"durationMs"`
}

type LanguageInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Template string `json:"template"`
}

type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (e *LocalExecutor) SupportedLanguages() []LanguageInfo {
	return []LanguageInfo{
		{
			ID:      "python",
			Name:    "Python 3",
			Version: "3.11",
			Template: "import sys\n\ndef solve():\n    n = int(input())\n    arr = list(map(int, input().split()))\n    # Your solution here\n    print(arr)\n\nsolve()\n",
		},
		{
			ID:      "cpp",
			Name:    "C++",
			Version: "C++17",
			Template: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n    int n;\n    cin >> n;\n    vector<int> arr(n);\n    for(int i=0;i<n;i++) cin >> arr[i];\n    // Your solution here\n    return 0;\n}\n",
		},
		{
			ID:      "java",
			Name:    "Java",
			Version: "21",
			Template: "import java.util.Scanner;\n\npublic class Main {\n    public static void main(String[] args) {\n        Scanner sc = new Scanner(System.in);\n        int n = sc.nextInt();\n        int[] arr = new int[n];\n        for(int i=0;i<n;i++) arr[i] = sc.nextInt();\n        // Your solution here\n    }\n}\n",
		},
		{
			ID:      "javascript",
			Name:    "JavaScript",
			Version: "Node.js 20",
			Template: "const lines = require('fs').readFileSync('/dev/stdin','utf8').trim().split('\\n');\nconst n = parseInt(lines[0]);\nconst arr = lines[1].split(' ').map(Number);\n// Your solution here\n",
		},
		{
			ID:      "go",
			Name:    "Go",
			Version: "1.21",
			Template: "package main\n\nimport (\n\t\"bufio\"\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\treader := bufio.NewReader(os.Stdin)\n\tvar n int\n\tfmt.Fscan(reader, &n)\n\tarr := make([]int, n)\n\tfor i := 0; i < n; i++ {\n\t\tfmt.Fscan(reader, &arr[i])\n\t}\n\t// Your solution here\n}\n",
		},
	}
}

func normalizeInput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func compile(args []string, timeoutSec int) (string, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(timeoutSec)*time.Second,
	)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (e *LocalExecutor) Run(code, language, input string, timeLimitMs int) (ExecutionResult, error) {
	log.Printf("[EXECUTOR] lang=%s input_len=%d", language, len(input))

	if timeLimitMs <= 0 {
		timeLimitMs = 2000
	}

	tmpDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return ExecutionResult{
			Error:     "failed to create temp dir",
			ErrorType: "internal_error",
		}, err
	}
	defer os.RemoveAll(tmpDir)

	// Normalize input — preserve ALL newlines, only fix \r\n
	input = normalizeInput(input)
	log.Printf("[EXECUTOR] normalized input=%q", input)

	var runArgs []string
	var runDir string

	switch strings.ToLower(strings.TrimSpace(language)) {
	case "python", "python3":
		srcPath := filepath.Join(tmpDir, "solution.py")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{
				Error:     "failed to write source",
				ErrorType: "internal_error",
			}, err
		}
		runArgs = []string{"python3", srcPath}

	case "cpp", "c++", "c":
		srcPath := filepath.Join(tmpDir, "solution.cpp")
		binPath := filepath.Join(tmpDir, "solution")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{
				Error:     "failed to write source",
				ErrorType: "internal_error",
			}, err
		}

		compOut, compErr := compile(
			[]string{"g++", "-O2", "-o", binPath, srcPath, "-std=c++17"},
			15,
		)
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
		className := "Main"
		re := regexp.MustCompile(`public\s+class\s+(\w+)`)
		if m := re.FindStringSubmatch(code); len(m) > 1 {
			className = m[1]
		}

		srcPath := filepath.Join(tmpDir, className+".java")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{
				Error:     "failed to write source",
				ErrorType: "internal_error",
			}, err
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
			return ExecutionResult{
				Error:     "failed to write source",
				ErrorType: "internal_error",
			}, err
		}
		runArgs = []string{"node", srcPath}

	case "go", "golang":
		srcPath := filepath.Join(tmpDir, "main.go")
		if err := os.WriteFile(srcPath, []byte(code), 0644); err != nil {
			return ExecutionResult{
				Error:     "failed to write source",
				ErrorType: "internal_error",
			}, err
		}
		// go run handles single-file compilation with no go.mod needed
		runArgs = []string{"go", "run", srcPath}
		runDir = tmpDir
		// Add 8s compile buffer — does NOT count toward user TLE verdict
		timeLimitMs += 8000

	default:
		return ExecutionResult{
			Error:     fmt.Sprintf("unsupported language: %s", language),
			ErrorType: "unsupported_language",
		}, fmt.Errorf("unsupported language: %s", language)
	}

	// ── Run phase ──
	timeout := time.Duration(timeLimitMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, runArgs[0], runArgs[1:]...)
	if runDir != "" {
		cmd.Dir = runDir
	}

	// Pipe input to stdin — newlines preserved exactly
	cmd.Stdin = strings.NewReader(input)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	startTime := time.Now()
	runErr := cmd.Run()
	duration := time.Since(startTime).Milliseconds()

	stdout := normalizeOutput(stdoutBuf.String())
	stderr := stderrBuf.String()

	log.Printf("[EXECUTOR] lang=%s duration=%dms stdout=%q stderr=%q",
		language, duration, stdout, stderr)

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

	if runErr != nil {
		exitCode := 0
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

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

	return ExecutionResult{
		Output:     stdout,
		ExitCode:   0,
		DurationMs: duration,
	}, nil
}
