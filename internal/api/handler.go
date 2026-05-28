package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/queue"
	"github.com/thesouldev/goboxd/internal/runner"
)

var cfg *config.Config
var jobQueue *queue.Queue

func Init(c *config.Config, q *queue.Queue) {
	cfg = c
	jobQueue = q
}

type Limits struct {
	WallTimeS    int `json:"wall_time_s"`
	MemoryKB     int `json:"memory_kb"`
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

type BuildResult struct {
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
}

type RunResponse struct {
	Status string       `json:"status"`
	Build  *BuildResult `json:"build,omitempty"`
	Tests  []TestResult `json:"tests"`
}

func validateFilename(name string) bool {
	if name == "" {
		return true
	}
	if strings.ContainsAny(name, "/\\") {
		return false
	}
	if strings.HasPrefix(name, ".") {
		return false
	}
	if len(name) > 255 {
		return false
	}
	return true
}

func isFlagAllowed(flag string, allowlist []string) bool {
	for _, pattern := range allowlist {
		if strings.HasSuffix(pattern, "*") {
			if strings.HasPrefix(flag, pattern[:len(pattern)-1]) {
				return true
			}
		} else if flag == pattern {
			return true
		}
	}
	return false
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

func RunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method_not_allowed", "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", err.Error())
		return
	}

	if req.Language == "" {
		writeError(w, 400, "missing_language", "language is required")
		return
	}

	if len(req.Tests) == 0 {
		writeError(w, 400, "missing_tests", "at least one test is required")
		return
	}

	if !validateFilename(req.SourceFilename) {
		writeError(w, 400, "invalid_filename", "source_filename must be a single path component")
		return
	}

	if !validateFilename(req.ArtifactFilename) {
		writeError(w, 400, "invalid_filename", "artifact_filename must be a single path component")
		return
	}

	lang := cfg.FindLanguage(req.Language)
	if lang == nil {
		writeError(w, 400, "unknown_language", "language not supported: "+req.Language)
		return
	}

	if len(req.Build.Flags) > 0 && len(lang.Build.FlagAllowlist) > 0 {
		for _, flag := range req.Build.Flags {
			if !isFlagAllowed(flag, lang.Build.FlagAllowlist) {
				writeError(w, 400, "disallowed_flag", "flag not allowed: "+flag)
				return
			}
		}
	}

	tests := make([]struct {
		Stdin          string
		ExpectedStdout string
	}, len(req.Tests))
	for i, t := range req.Tests {
		tests[i].Stdin = t.Stdin
		tests[i].ExpectedStdout = t.ExpectedStdout
	}

	// Run job through queue — waits if server is busy
	var jobResult runner.JobResult
	jobQueue.Run(func() {
		jobResult = runner.RunJob(lang, req.Source, req.Build.Flags, tests)
	})

	resp := RunResponse{}

	if jobResult.Build != nil {
		status := "ok"
		if jobResult.Build.ExitCode != 0 {
			status = "failed"
		}
		resp.Build = &BuildResult{
			Status:     status,
			Stdout:     jobResult.Build.Stdout,
			Stderr:     jobResult.Build.Stderr,
			DurationMs: jobResult.Build.DurationMs,
		}
		if status == "failed" {
			resp.Status = "build_failed"
			resp.Tests = make([]TestResult, len(req.Tests))
			for i := range resp.Tests {
				resp.Tests[i] = TestResult{Status: "not_executed"}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	overallStatus := "accepted"
	for i, result := range jobResult.Tests {
		got := strings.TrimRight(result.Stdout, "\n")
		expected := strings.TrimRight(req.Tests[i].ExpectedStdout, "\n")

		status := "accepted"
		if result.ExitCode != 0 {
			status = "runtime_error"
		} else if got == expected {
			status = "accepted"
		} else if strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(expected)) {
			status = "output_whitespace_mismatch"
		} else {
			status = "wrong_output"
		}

		if status != "accepted" && overallStatus == "accepted" {
			overallStatus = status
		}

		resp.Tests = append(resp.Tests, TestResult{
			Status:     status,
			Stdout:     result.Stdout,
			Stderr:     result.Stderr,
			DurationMs: result.DurationMs,
		})
	}

	resp.Status = overallStatus
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}