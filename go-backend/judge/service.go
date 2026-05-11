package judge

import (
	"log"
	"strings"
	"sync"
)

// ──────────────────────────────────────────────
// ExecutionService — singleton service layer.
//
// Architecture:
//   frontend → backend API/ws → ExecutionService → Executor → result
//
// The ExecutionService wraps an Executor with:
//   - Concurrency limiting (maxConcurrent)
//   - Language metadata caching
//   - Future: queue, retry, metrics, sandboxing
//
// To switch executor backends (e.g., Docker, cloud worker):
//   1. Implement the Executor interface
//   2. Change newDefaultExecutor() below
// ──────────────────────────────────────────────

var (
	service     *ExecutionService
	serviceOnce sync.Once
)

// ExecutionService manages code execution lifecycle.
type ExecutionService struct {
	executor      Executor
	semaphore     chan struct{} // limits concurrency
	maxConcurrent int
}

// GetService returns the singleton ExecutionService.
func GetService() *ExecutionService {
	serviceOnce.Do(func() {
		service = &ExecutionService{
			executor:      newDefaultExecutor(),
			maxConcurrent: 10, // max parallel executions
			semaphore:     make(chan struct{}, 10),
		}
	})
	return service
}

// newDefaultExecutor returns the executor to use.
// SWAP THIS to switch from local → Docker → cloud.
func newDefaultExecutor() Executor {
	return NewLocalExecutor()
}

// Execute runs code through the executor with concurrency control.
func (s *ExecutionService) Execute(code, language, input string, timeLimitMs int) (ExecutionResult, error) {
	log.Printf("[SERVICE] running language: %s", language)
	
	// Acquire semaphore slot
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	return s.executor.Run(code, language, input, timeLimitMs)
}

// GetLanguages returns metadata about supported languages.
func (s *ExecutionService) GetLanguages() []LanguageInfo {
	return s.executor.SupportedLanguages()
}

// RunTestCases executes code against multiple inputs and returns results.
// This is a convenience method for batch execution.
func (s *ExecutionService) RunTestCases(code, language string, inputs []string, expectedOutputs []string, timeLimitMs int) []TestCaseResult {
	results := make([]TestCaseResult, len(inputs))
	for i, inp := range inputs {
		res, err := s.Execute(code, language, inp, timeLimitMs)
		
		tcr := TestCaseResult{
			Input:      inp,
			Expected:   expectedOutputs[i],
			Actual:     res.Output,
			ExecResult: res,
		}

		if err != nil {
			tcr.Actual = "Execution error: " + err.Error()
			tcr.Pass = false
		} else if res.Error != "" {
			tcr.Pass = false
		} else {
			tcr.Pass = Normalize(tcr.Actual) == Normalize(tcr.Expected)
		}

		results[i] = tcr
		log.Printf("[JUDGE] case %d lang=%s pass=%v\nACTUAL:   %q\nEXPECTED: %q",
			i, language, tcr.Pass, Normalize(tcr.Actual), Normalize(tcr.Expected))
	}
	return results
}

// TestCaseResult is the result of running code against one test case.
type TestCaseResult struct {
	Input      string          `json:"input"`
	Expected   string          `json:"expected"`
	Actual     string          `json:"actual"`
	Pass       bool            `json:"pass"`
	ExecResult ExecutionResult `json:"execResult"`
}



// Normalize prepares a string for comparison:
// 1. Replace \r\n with \n and strip stray \r
// 2. Trim trailing whitespace from each line
// 3. Remove trailing empty lines
func Normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	// Remove trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
