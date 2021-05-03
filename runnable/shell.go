package runnable

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"github.com/pkg/errors"
)

// ShellCommand executes a given command using a shell.
//
type ShellCommand struct {
	command string
	concourseUrl string
	stderr  io.Writer
	stdout  io.Writer
}

var _ Runnable = &ShellCommand{}

// NewShellCommand instantiates a new ShellCommand that
// is meant to always run with a specified `command`.
//
func NewShellCommand(command string, stderr io.Writer) (runnable *ShellCommand) {
	runnable = &ShellCommand{
		command: command,
		stderr:  stderr,
	}
	return
}

// Run runs the command either until completion or context
// cancellation.
//
func (r *ShellCommand) Run(ctx context.Context) (err error) {
	var output []byte

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", r.command)
	output, err = cmd.CombinedOutput()
	//fmt.Println("=======Command output for each function =======")
	//fmt.Println(string(output))
  //fmt.Println("=======Command output for each function =======")
	switch ctx.Err() {
	case context.DeadlineExceeded:
		err = errors.Wrapf(ctx.Err(),
			"command didn't finish on time")
		 fmt.Fprintf(r.stderr, "COMMAND DIDN'T FINISH IN TIME --- \n%s\n", output)
	case context.Canceled:
		err = errors.Wrapf(ctx.Err(),
			"command execution cancelled")
		  fmt.Fprintf(r.stderr, "COMMAND EXECUTION WAS CANCELLED --- \n%s\n", output)
	}

	if err != nil {
		err = errors.Wrapf(err,
			"command execution failed")
		  fmt.Fprintf(r.stderr, "COMMAND FAILURE--- \n%s\n", output)
		return
	}
	return
}
