package main

import (
	"io/fs"
	"strings"
)

func quotedString(input string) string {
	if ContainsWhitespace(input) {
		return `"` + input + `"`
	}
	return input
}

func makeCommandLine(cmd string, params []string) string {
	parts := make([]string, 0)
	parts = append(parts, quotedString(cmd))
	for _, param := range params {
		parts = append(parts, quotedString(param))
	}
	return strings.Join(parts, " ")
}

func executeArgs(cmd string, params []string, path string, info fs.FileInfo) {
	if ARG_DEBUG {
		str := makeCommandLine(cmd, params)
		println(str)
		return
	}

	_, exsits := INNER_COMMAND[cmd]
	if exsits {
		executeInnerCommand(cmd, params, path, info)
	} else {
		exec := NewCommandExecution(ExecutiionConext{
			CommandName: "command",
			CommandLine: makeCommandLine(cmd, params),
			Workdir:     ".",
			Params:      EMPTY_PARAMS,
		}, ExecutionCallbacks{})
		exec.Execute()
	}
}

func replaceArgParam(param string, dirPath string, filepath string) string {
	switch param {
	case "{}":
		return filepath
	case "[]":
		return dirPath
	default:
		return param
	}
}

func preReplaceExecArgs(args [][]string, path string, info fs.FileInfo) [][]string {
	dirPath := getDirPath(path, info)
	_, exists := PATH_MAP[dirPath]
	PATH_MAP[dirPath] = true

	if !info.IsDir() {
		FILE_LIST = append(FILE_LIST, path)
	} else {
		DIR_LIST = append(DIR_LIST, path)
	}

	// No need to process
	if !WILL_PROCESS_DIR && !WILL_PROCESS_FILE {
		return args
	}

	execDir := false
	if WILL_PROCESS_DIR {
		execDir = !exists
	}

	// Replace parameters
	newArgs := make([][]string, 0)
	for _, arg := range args {
		dir, _ := getExecParamType(arg[1:])
		if dir {
			if !execDir {
				continue
			}
		}
		argParams := make([]string, 0)
		for _, param := range arg {
			argParams = append(argParams, replaceArgParam(param, dirPath, path))
		}
		newArgs = append(newArgs, argParams)

	}
	return newArgs
}
