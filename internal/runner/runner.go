package runner

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
)

type Result struct {
	Stdout     string
	Stderr     string
	DurationMs int64
	ExitCode   int
}

type JobResult struct {
	Build *Result
	Tests []Result
}

func replaceTemplate(s, source, artifact string) string {
	s = strings.ReplaceAll(s, "{{source}}", source)
	s = strings.ReplaceAll(s, "{{artifact}}", artifact)
	return s
}

func runDirect(cmd string, args []string, stdin string, dir string, wallTimeS int) Result {
	if wallTimeS <= 0 {
		wallTimeS = 10
	}

	c := exec.Command(cmd, args...)
	c.Stdin = bytes.NewBufferString(stdin)
	c.Dir = dir

	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf

	done := make(chan error, 1)
	start := time.Now()

	if err := c.Start(); err != nil {
		return Result{Stderr: err.Error(), ExitCode: 1}
	}

	go func() { done <- c.Wait() }()

	select {
	case <-time.After(time.Duration(wallTimeS+5) * time.Second):
		c.Process.Kill()
		return Result{
			Stderr:     "time limit exceeded",
			DurationMs: time.Since(start).Milliseconds(),
			ExitCode:   124,
		}
	case err := <-done:
		duration := time.Since(start).Milliseconds()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}
		out := outBuf.String()
		if len(out) > 4*1024*1024 {
			out = out[:4*1024*1024] + "\n[output truncated]"
		}
		return Result{
			Stdout:     out,
			Stderr:     errBuf.String(),
			DurationMs: duration,
			ExitCode:   exitCode,
		}
	}
}

func RunJob(lang *config.Language, source string, flags []string, tests []struct {
	Stdin          string
	ExpectedStdout string
}) JobResult {
	jailDir, err := os.MkdirTemp("/tmp", "job-*")
	if err != nil {
		return JobResult{Build: &Result{Stderr: "failed to create temp dir", ExitCode: 1}}
	}
	defer os.RemoveAll(jailDir)
	os.Chmod(jailDir, 0777)

	sourceFilename := lang.SourceFilename
	if sourceFilename == "" {
		sourceFilename = "solution.txt"
	}
	sourcePath := filepath.Join(jailDir, sourceFilename)
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		return JobResult{Build: &Result{Stderr: "failed to write source", ExitCode: 1}}
	}
	os.Chmod(sourcePath, 0644)

	artifactPath := ""
	if lang.Artifact != "" {
		artifactPath = filepath.Join(jailDir, lang.Artifact)
	}

	if lang.Build.Cmd != "" {
		args := make([]string, len(lang.Build.Args))
		for i, a := range lang.Build.Args {
			args[i] = replaceTemplate(a, sourcePath, artifactPath)
		}
		args = append(args, flags...)

		wallTime := lang.Build.Limits.WallTimeS
		if wallTime == 0 {
			wallTime = 10
		}

		buildResult := runDirect(lang.Build.Cmd, args, "", jailDir, wallTime)
		if buildResult.ExitCode != 0 {
			return JobResult{Build: &buildResult}
		}

		if artifactPath != "" {
			os.Chmod(artifactPath, 0755)
		}

		return runTests(lang, sourcePath, artifactPath, jailDir, tests, &buildResult)
	}

	return runTests(lang, sourcePath, artifactPath, jailDir, tests, nil)
}

func runTests(lang *config.Language, sourcePath, artifactPath, jailDir string, tests []struct {
	Stdin          string
	ExpectedStdout string
}, buildResult *Result) JobResult {
	job := JobResult{Build: buildResult}

	wallTime := lang.Run.Limits.WallTimeS
	if wallTime == 0 {
		wallTime = 10
	}

	for _, test := range tests {
		args := make([]string, len(lang.Run.Args))
		for i, a := range lang.Run.Args {
			args[i] = replaceTemplate(a, sourcePath, artifactPath)
		}

		cmd := replaceTemplate(lang.Run.Cmd, sourcePath, artifactPath)
		result := runDirect(cmd, args, test.Stdin, jailDir, wallTime)
		job.Tests = append(job.Tests, result)
	}

	return job
}