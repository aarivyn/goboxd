package runner

import (
	"bytes"
	"fmt"
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
	MemPeakKB  int64
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

func runCmd(cmd string, args []string, stdin string, dir string, timeLimitSec int) Result {
nsjailArgs := []string{
    "--mode", "o",
    "--time_limit", fmt.Sprintf("%d", timeLimitSec),
    "--rlimit_as", "512",
    "--rlimit_nproc", "64",
    "--rlimit_fsize", "32",
    "--user", "65534",
    "--group", "65534",
    "--chroot", "/",
    "--cwd", dir,
    "--bindmount", dir + ":" + dir,
    "--bindmount_ro", "/usr:/usr",
    "--bindmount_ro", "/lib:/lib",
    "--bindmount_ro", "/lib64:/lib64",
    "--bindmount_ro", "/bin:/bin",
    "--tmpfsmount", "/etc",
    "--tmpfsmount", "/tmp",
    "--disable_clone_newnet",
    "--",
    cmd,
}
	nsjailArgs = append(nsjailArgs, args...)

	c := exec.Command("/usr/sbin/nsjail", nsjailArgs...)
	c.Stdin = bytes.NewBufferString(stdin)

	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf

	start := time.Now()
	err := c.Run()
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

func RunJob(lang *config.Language, source string, flags []string, tests []struct {
	Stdin          string
	ExpectedStdout string
}) JobResult {
	jailDir, err := os.MkdirTemp("/var/tmp", "job-*")
	if err != nil {
		return JobResult{Build: &Result{Stderr: "failed to create temp dir", ExitCode: 1}}
	}
	defer os.RemoveAll(jailDir)

	if err := os.MkdirAll(jailDir, 0755); err != nil {
		return JobResult{Build: &Result{Stderr: "failed to prepare jail dir", ExitCode: 1}}
	}

	sourceFilename := lang.SourceFilename
	if sourceFilename == "" {
		sourceFilename = "solution.txt"
	}
	sourcePath := filepath.Join(jailDir, sourceFilename)
	if err := os.WriteFile(sourcePath, []byte(source), 0755); err != nil {
		return JobResult{Build: &Result{Stderr: "failed to write source", ExitCode: 1}}
	}

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

		buildTimeLimit := lang.Build.TimeLimit
		if buildTimeLimit <= 0 {
			buildTimeLimit = 30
		}

		buildResult := runCmd(lang.Build.Cmd, args, "", jailDir, buildTimeLimit)
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

	runTimeLimit := lang.Run.TimeLimit
	if runTimeLimit <= 0 {
		runTimeLimit = 10
	}

	for _, test := range tests {
		args := make([]string, len(lang.Run.Args))
		for i, a := range lang.Run.Args {
			args[i] = replaceTemplate(a, sourcePath, artifactPath)
		}

		cmd := replaceTemplate(lang.Run.Cmd, sourcePath, artifactPath)

		result := runCmd(cmd, args, test.Stdin, jailDir, runTimeLimit)
		job.Tests = append(job.Tests, result)
	}

	return job
}