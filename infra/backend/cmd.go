package backend

import (
	"bytes"
	"fmt"
	"github.com/kohkimakimoto/cofu/infra/util"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

type Cmd struct {
	Shell string
}

func NewCmd(shell string) *Cmd {
	return &Cmd{
		Shell: shell,
	}
}

func (c *Cmd) BuildCommand(command string, option *CommandOption) string {
	if option != nil {
		if option.Cwd != "" {
			command = fmt.Sprintf("cd %s && %s", util.ShellEscape(option.Cwd), command)
		}

		if option.User != "" {
			command = fmt.Sprintf("cd ~%s && %s", option.User, command)
			command = fmt.Sprintf("sudo -H -u %s -- %s -c %s", util.ShellEscape(option.User), util.ShellEscape(c.Shell), util.ShellEscape(command))
		}
	}

	return command
}

func (c *Cmd) RunCommand(command string) *CommandResult {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command(c.Shell, "-c", command)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var combined bytes.Buffer

	cmd.Stdout = io.MultiWriter(&stdout, &combined)
	cmd.Stderr = io.MultiWriter(&stderr, &combined)
	cmd.Stdin = os.Stdin

	var exitStatus int
	err := cmd.Run()
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			} else {
				panic("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus.")
			}
		}
	} else {
		exitStatus = 0
	}

	return &CommandResult{
		Stdout:     stdout,
		Stderr:     stderr,
		Combined:   combined,
		Err:        err,
		ExitStatus: exitStatus,
	}
}

type CommandResult struct {
	Stdout     bytes.Buffer
	Stderr     bytes.Buffer
	Combined   bytes.Buffer
	ExitStatus int
	Err        error
}

func (r *CommandResult) Success() bool {
	return r.ExitStatus == 0
}

func (r *CommandResult) Failure() bool {
	return r.ExitStatus != 0
}

type CommandOption struct {
	User string
	Cwd  string
}
