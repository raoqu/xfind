package main

import (
	"bufio"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type ExectuionOutputMatchCallback func(line string) bool
type ExecutionOutputCallback func(matchedLine string)
type ExecutionEventCallback func(event string, param string)

type ExecutiionConext struct {
	CommandName string
	CommandLine string
	Workdir     string
	Params      []string
}

type ExecutionCallbacks struct {
	MatchOutputCallback ExectuionOutputMatchCallback
	OutputCallback      ExecutionOutputCallback
	EventCallback       ExecutionEventCallback
}

type CommandExecution struct {
	CommandName string
	CommandLine string
	Workdir     string
	Params      []string

	MatchOutputCallback ExectuionOutputMatchCallback
	OutputCallback      ExecutionOutputCallback
	EventCallback       ExecutionEventCallback
}

// NewApp creates a new App application struct
func NewCommandExecution(context ExecutiionConext, callbacks ExecutionCallbacks) *CommandExecution {
	return &CommandExecution{
		CommandName: context.CommandName,
		CommandLine: context.CommandLine,
		Workdir:     context.Workdir,
		Params:      context.Params,

		MatchOutputCallback: callbacks.MatchOutputCallback,
		OutputCallback:      callbacks.OutputCallback,
		EventCallback:       callbacks.EventCallback,
	}
}

func (c *CommandExecution) Execute() string {

	parts := strings.Fields(c.CommandLine)
	if len(parts) == 0 {
		return ""
	}

	cmd := buildExecCmd(c.Workdir, parts, c.Params)

	// cmdLine := strings.Join(cmd.Args, " ")

	if len(c.Workdir) > 0 {
		cmd.Dir = c.Workdir
		c.event("shell_info", c.Workdir)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}

	if err := cmd.Start(); err != nil {
		c.event("shell_error", c.CommandName)
		return ""
	}

	// c.event("shell", "COMMAND LINE: "+cmdLine)

	var pid = cmd.Process.Pid

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			c.output(line)
		}

		c.event("shell_complete", c.CommandName)
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return ""
	}

	return strconv.Itoa(pid)
}

func (c *CommandExecution) output(line string) {
	if c.MatchOutputCallback != nil && c.MatchOutputCallback(line) {
		if c.OutputCallback != nil {
			c.OutputCallback(line)
		}
	}
}

func (c *CommandExecution) event(evt string, param string) {
	if c.EventCallback != nil {
		c.EventCallback(evt, param)
	}
}

func buildExecCmd(dir string, parts []string, params []string) *exec.Cmd {
	shell := OSFeature("c:\\windows\\system32\\cmd.exe", "/bin/sh")
	flag := OSFeature("/C", "-c")

	command := makeShellCommand(parts, params)
	cmd := exec.Command(shell, flag, command)

	initCmdSysProcAttr(cmd)

	if len(dir) > 0 {
		cmd.Dir = dir
	}

	return cmd
}

func OSFeature(winFeature, linuxFeature string) string {
	if runtime.GOOS == "windows" {
		return winFeature
	}

	return linuxFeature
}

/*
$1 -> params[0]
$2 -> params[1] ...
*/
func makeShellCommand(parts []string, params []string) string {
	result := make([]string, len(parts))
	copy(result, parts)

	for i, part := range result {
		if strings.HasPrefix(part, "$") {
			indexStr := part[1:]
			index, err := strconv.Atoi(indexStr)
			if err == nil && index > 0 && index <= len(params) {
				result[i] = params[index-1]
			}
		}
		result[i] = quoteIfSpace(result[i])
	}

	return strings.Join(result, " ")
}

func quoteIfSpace(s string) string {
	if strings.Contains(s, " ") {
		if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
			return "\"" + s + "\""
		}
	}
	return s
}
