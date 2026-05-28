package runner

import (
	"bytes"
	"os"
	"os/exec"
	"time"
)

type Result struct {
	Stdout     string
	Stderr     string
	DurationMs int64
	ExitCode   int
}

func Run(cmd string, args []string, stdin string) Result {
	// Create a unique temp folder for this job
	jailDir, err := os.MkdirTemp("/tmp", "job-*")
	if err != nil {
		return Result{Stderr: "failed to create temp dir", ExitCode: 1}
	}
	// Always clean up when done - even if something crashes
	defer os.RemoveAll(jailDir)

	// Build the command
	c := exec.Command(cmd, args...)
	c.Stdin = bytes.NewBufferString(stdin)
	c.Dir = jailDir

	var outBuf, errBuf bytes.Buffer
	c.Stdout = &outBuf
	c.Stderr = &errBuf

	// Run it and measure time
	start := time.Now()
	err = c.Run()
	duration := time.Since(start).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return Result{
		Stdout:     outBuf.String(),
		Stderr:     errBuf.String(),
		DurationMs: duration,
		ExitCode:   exitCode,
	}
}