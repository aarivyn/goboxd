 package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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
	MemoryPeakKB int64 `json:"memory_peak_kb"`
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
	json.NewEncoder(w).Encode(map[string]any{
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

	// Security: cap request body at 256KB
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "bad_request", err.Error())
		return
	}

	// Structured request log 
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
        startTime := time.Now()
	log.Printf(`{"request_id":"%s","language":"%s","event":"received"}`,
		requestID, req.Language)

	if req.Language == "" {
		writeError(w, 400, "missing_language", "language is required")
		return
	}

	if len(req.Tests) == 0 {
		writeError(w, 400, "missing_tests", "at least one test is required")
		return
	}

	// Security: validate filenames — no path traversal
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

	// Security: flag allowlist — no compiler flag injection
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

	// Run through concurrency queue — waits if server is busy, never fails
	
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
			log.Printf(`{"request_id":"%s","language":"%s","status":"build_failed","duration_ms":%d}`,
				requestID, req.Language, time.Since(startTime).Milliseconds())
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
			MemoryPeakKB: result.MemoryPeakKB,
		})
	}

		resp.Status = overallStatus

	// Record metrics
	durationMs := time.Since(startTime).Milliseconds()
	RecordRequest(req.Language, resp.Status, durationMs)

	// Structured completion log
	log.Printf(`{"request_id":"%s","language":"%s","status":"%s","duration_ms":%d}`,
		requestID, req.Language, resp.Status, time.Since(startTime).Milliseconds())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LanguagesHandler handles both /languages (list all) and /languages/{id} (get one)
func LanguagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "method_not_allowed", "method not allowed")
		return
	}

	// Parse the path to check if a specific language ID is requested
	// /languages -> list all
	// /languages/{id} -> get specific language
	path := strings.TrimPrefix(r.URL.Path, "/languages")
	path = strings.TrimPrefix(path, "/")

	w.Header().Set("Content-Type", "application/json")

	if path == "" {
		// List all languages
		listLanguages(w, r)
	} else {
		// Get specific language detail
		getLanguageDetail(w, r, path)
	}
}

func listLanguages(w http.ResponseWriter, r *http.Request) {
	type LanguageItem struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type ListResponse struct {
		Languages []LanguageItem `json:"languages"`
	}

	languages := make([]LanguageItem, 0, len(cfg.Languages))
	for _, lang := range cfg.Languages {
		languages = append(languages, LanguageItem{
			ID:   lang.ID,
			Name: lang.Name,
		})
	}

	resp := ListResponse{Languages: languages}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func getLanguageDetail(w http.ResponseWriter, r *http.Request, langID string) {
	type DetailResponse struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Compiled bool   `json:"compiled"`
		Extension string `json:"extension"`
	}

	lang := cfg.FindLanguage(langID)
	if lang == nil {
		writeError(w, 404, "not_found", "language not found: "+langID)
		return
	}

	// Determine if compiled: has a Build command defined
	compiled := lang.Build.Cmd != ""

	// Extract file extension from SourceFilename
	ext := filepath.Ext(lang.SourceFilename)
	if ext == "" {
		ext = ""
	}

	resp := DetailResponse{
		ID:        lang.ID,
		Name:      lang.Name,
		Compiled:  compiled,
		Extension: ext,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}