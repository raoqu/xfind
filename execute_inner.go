package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

var SWITCH_COUNT_FILES = false
var SWITCH_COUNT_MATCHES = false
var SWITCH_COUNT_LINES = false
var SWITCH_PRINT_MATCHES = false

var COUNTER_ALL_FILES = 0
var COUNTER_MATCH_LINES = 0
var COUNTER_ALL_LINES = 0

func parseInnerCommand(cmd string, _ []string) {
	switch cmd {
	case CMD_PRINT:
		// default no action
	case CMD_PRINT_MATCHES:
		SWITCH_PRINT_MATCHES = true
	case CMD_COUNT_FILES:
		SWITCH_COUNT_FILES = true
	case CMD_COUNT_LINES:
		SWITCH_COUNT_LINES = true
	case CMD_COUNT_MATCHES:
		SWITCH_COUNT_MATCHES = true
	default:
		println("Inner command '", cmd, "' parser NOT IMPLEMENTED")
	}
}

func executeInnerCommand(cmd string, params []string, path string, info fs.FileInfo) {
	switch cmd {
	case CMD_PRINT:
		executePrint(params)
	case CMD_COUNT_FILES:
		COUNTER_ALL_FILES++
	case CMD_PRINT_MATCHES:
	case CMD_COUNT_LINES:
	case CMD_COUNT_MATCHES:
		// will execute in 'defaultInnerCommand'
	default:
		println("Inner command '", cmd, "' executor NOT IMPLEMENTED")
	}
}

// after all inner command executed
func postInnerCommandExecution() {
	if SWITCH_COUNT_FILES {
		println("file total:", COUNTER_ALL_FILES)
	}
	if SWITCH_COUNT_LINES {
		println("line total:", COUNTER_ALL_LINES)
	}
	if SWITCH_COUNT_MATCHES {
		println("line matches:", COUNTER_MATCH_LINES)
	}

	if ARG_DELETE {
		executeDelete()
	}
}

func defaultInnerCommand(path string, info fs.FileInfo) {
	if ARG_DEBUG {
		return
	}

	if SWITCH_COUNT_MATCHES || SWITCH_PRINT_MATCHES || SWITCH_COUNT_LINES {
		executeParseFile(path, info)
	}
}

func executePrint(params []string) {
	if len(params) > 0 {
		str := strings.Join(params, " ")
		fmt.Println(str)
	}
}

func matchRegex(text string, matched *string) bool {
	for _, regex := range ARG_REGEX {
		matches := regex.FindStringSubmatch(text)
		if matches != nil {
			// matches[0] == text
			if len(matches) == 1 {
				*matched = matches[0]
			} else {
				*matched = matches[1]
			}
			return true
		}
	}
	return false
}

func matchText(text string) bool {
	for _, m := range ARG_MATCHES {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

func executeParseFile(path string, info fs.FileInfo) {
	if len(ARG_MATCHES) < 1 && len(ARG_REGEX) < 1 && !SWITCH_COUNT_LINES {
		return
	}

	if !info.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("printmatch: ERROR opening file:", path)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1

		for scanner.Scan() {
			text := scanner.Text()
			if SWITCH_COUNT_LINES {
				COUNTER_ALL_LINES++
			}

			var matched = text
			if matchText(text) || matchRegex(text, &matched) {
				// print line matches
				if SWITCH_PRINT_MATCHES {
					fmt.Printf("%5d: %s\n", lineNumber, matched)
				}
				// count line matches
				if SWITCH_COUNT_MATCHES {
					COUNTER_MATCH_LINES++
				}
			}
			lineNumber++
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("printmatch: ERROR reading file:", err)
		}
	}
}

func executeDelete() {
	for _, f := range FILE_LIST {
		fmt.Println("Delete file ", f)
		if !ARG_DEBUG {
			doDelete(f, false)
		}
	}
	for _, d := range DIR_LIST {
		fmt.Println("Delete folder ", d)
		if !ARG_DEBUG {
			doDelete(d, true)
		}
	}
}

func doDelete(path string, isdir bool) {
	if isdir {
		err := os.RemoveAll(path)
		if err != nil {
			println(err.Error())
		}
	} else {
		err := os.Remove(path)
		if err != nil {
			println(err.Error())
		}
	}
}
