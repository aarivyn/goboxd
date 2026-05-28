package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thesouldev/goboxd/internal/runner"
)

type Limits struct {
	WallTimeS   int `json:"wall_time_s"`
	MemoryKB    int `json:"memory_kb"`
	MaxProcesses int `json:"max_processes"`
}

type BuildRun struct {
	Limits Limits   `json:"limits"`
	Flags  []string `json:"flags"`
}

type Test struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

type RunRequest struct {
	Language         string   `json:"language"`
	Source           string   `json:"source"`
	SourceFilename   string   `json:"source_filename"`
	ArtifactFilename string   `json:"artifact_filename"`
	Build            BuildRun `json:"build"`
	Run              BuildRun `json:"run"`
	Tests            []Test   `json:"tests"`
}

type TestResult struct {
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
}

type RunResponse struct {
	Status  string       `json:"status"`
	Tests   []TestResult `json:"tests"`
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request size to 256KB
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	// Parse the JSON body
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", err.Error())
		return
	}

	// Check language is provided
	if req.Language == "" {
		writeError(w, 400, "missing_language", "language is required")
		return
	}

	// Check we have at least one test
	if len(req.Tests) == 0 {
		writeError(w, 400, "missing_tests", "at least one test is required")
		return
	}

	// For now only support py3
	if req.Language != "py3" {
		writeError(w, 400, "unknown_language", "language not supported: "+req.Language)
		return
	}

	// Write source to temp file and run it
	var testResults []TestResult
	overallStatus := "accepted"

	for _, test := range req.Tests {
		result := runner.Run("python3", []string{"-c", req.Source}, test.Stdin)

		// Compare output
		got := strings.TrimRight(result.Stdout, "\n")
		expected := strings.TrimRight(test.ExpectedStdout, "\n")

		status := "accepted"
		if result.ExitCode != 0 {
			status = "runtime_error"
		} else if got != expected {
			status = "wrong_output"
		}

		if status != "accepted" && overallStatus == "accepted" {
			overallStatus = status
		}

		testResults = append(testResults, TestResult{
			Status:     status,
			Stdout:     result.Stdout,
			Stderr:     result.Stderr,
			DurationMs: result.DurationMs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RunResponse{
		Status: overallStatus,
		Tests:  testResults,
	})
}

func writeError(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    errCode,
			"message": msg,
		},
	})
}