package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
)

type ArgState int

type ArgInfo struct {
	State   ArgState
	MinArgs int
	MaxArgs int
}

// arguments
var STATE_MAP = map[string]ArgInfo{
	"-name":    {State: argStateName, MinArgs: 1, MaxArgs: 1},
	"-match":   {State: argStateMatch, MinArgs: 1, MaxArgs: 1},
	"-regex":   {State: argStateRegex, MinArgs: 1, MaxArgs: 1},
	"-exclude": {State: argStateExclude, MinArgs: 1, MaxArgs: 1},
	"-exec":    {State: argStateExec, MinArgs: 1, MaxArgs: 100},
	"-delete":  {State: argStateDelete, MinArgs: 0, MaxArgs: 0},
	"-type":    {State: argStateType, MinArgs: 1, MaxArgs: 1},
	"-debug":   {State: argStateDebug, MinArgs: 0, MaxArgs: 0},
}

const (
	argStateNone    = 1
	argStateError   = 2
	argStateName    = 3
	argStateExec    = 4
	argStateExclude = 5
	argStateMatch   = 6
	argStateDelete  = 7
	argStateType    = 8
	argStateDebug   = 9
	argStateRegex   = 10
)

var ARG_NAMES = make([]string, 0)
var ARG_MATCHES = make([]string, 0)
var ARG_REGEX = make([]*regexp.Regexp, 0)
var ARG_EXCLUDES = make([]string, 0)
var ARG_EXEC = make([][]string, 0)
var ARG_DELETE = false
var ARG_TYPE_DIR = false
var ARG_TYPE_FILE = false
var ARG_DEBUG = false

var CURRENT_PARAMS = make([]string, 0)
var CURRENT_ARG = ""
var CURRENT_STATE ArgState = argStateNone

// Inner commands
const CMD_PRINT = "print"
const CMD_PRINT_MATCHES = "printmatch"
const CMD_COUNT_FILES = "count"
const CMD_COUNT_LINES = "countlines"
const CMD_COUNT_MATCHES = "countmatch"

var INNER_COMMAND = map[string]bool{
	CMD_PRINT_MATCHES: true,
	CMD_PRINT:         true,
	CMD_COUNT_FILES:   true,
	CMD_COUNT_MATCHES: true,
	CMD_COUNT_LINES:   true,
}

func parseArgs(args []string) {
	for i, arg := range args {
		if CURRENT_STATE == argStateNone {
			CURRENT_STATE = parseState(arg)
			if CURRENT_STATE == argStateError {
				panic(fmt.Sprintf("INVALID ARGUMENT (%d): %s", i, arg))
			}
			CURRENT_ARG = arg
		} else {
			nstate := parseState(arg)
			if nstate == argStateError {
				CURRENT_PARAMS = append(CURRENT_PARAMS, arg)
			} else {
				checkArgParams(CURRENT_STATE)
				CURRENT_STATE = nstate
				CURRENT_PARAMS = make([]string, 0)
				CURRENT_ARG = arg
			}
		}
	}
	if CURRENT_STATE != argStateNone {
		checkArgParams(CURRENT_STATE)
		CURRENT_STATE = argStateNone
		CURRENT_PARAMS = make([]string, 0)
	}
}

func parseState(param string) ArgState {
	info, exists := STATE_MAP[param]
	if exists {
		return info.State
	}
	return argStateError
}

func validateParamCount(mincount, maxcount int) {
	count := len(CURRENT_PARAMS)
	if count < mincount || count > maxcount {
		panic("Invalid parameter count for argument '" + CURRENT_ARG + "'")
	}
}

func checkArgParams(state ArgState) {
	for _, info := range STATE_MAP {
		if info.State == state {
			validateParamCount(info.MinArgs, info.MaxArgs)
		}
	}

	switch state {
	case argStateName:
		ARG_NAMES = append(ARG_NAMES, CURRENT_PARAMS...)
	case argStateMatch:
		ARG_MATCHES = append(ARG_MATCHES, CURRENT_PARAMS...)
	case argStateRegex:
		for _, rx := range CURRENT_PARAMS {
			ARG_REGEX = append(ARG_REGEX, regexp.MustCompile(rx))
		}
	case argStateExclude:
		ARG_EXCLUDES = append(ARG_EXCLUDES, CURRENT_PARAMS...)
	case argStateExec:
		ARG_EXEC = append(ARG_EXEC, CURRENT_PARAMS)
	case argStateDelete:
		ARG_DELETE = true
	case argStateDebug:
		ARG_DEBUG = true
	case argStateType:
		ftype := CURRENT_PARAMS[0]
		if ftype == "file" {
			ARG_TYPE_FILE = true
		} else if ftype == "dir" {
			ARG_TYPE_DIR = true
		} else if ftype == "both" {
			ARG_TYPE_FILE = true
			ARG_TYPE_DIR = true
		} else {
			panic("Invalid parameter '" + ftype + "' for '-type'")
		}
	case argStateNone:
	case argStateError:
		// Nothting to do
	default:
		panic("Not implemented parameter type " + strconv.Itoa(int(state)))
	}
}

func parseInnerCommands() {
	for _, arg := range ARG_EXEC {
		cmd := arg[0]
		params := arg[1:]
		_, exists := INNER_COMMAND[cmd]
		if exists {
			parseInnerCommand(cmd, params)
		}
	}
}

func execParamContains(params []string, target string) bool {
	for _, param := range params {
		if param == target {
			return true
		}
	}
	return false
}

func getDirPath(path string, info fs.FileInfo) string {
	if info.IsDir() {
		return path
	} else {
		dir, _ := filepath.Split(path)
		return dir
	}
}
