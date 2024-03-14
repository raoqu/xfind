package main

/*
用法:
	xfind . -name "*.go" -match "Text Match" -delete
	xfind . -name "*.go" -exec print {}
	xfind . -name "*.go" -exec print []
	xfind . -name "*.go" -exclude "*.ioc.go"
	xfind . -name "*.go" -exclude ".git"
	xfind . -name "*.ioc.go" -delete
	xfind . -type dir -name ".git" -exec print [] -exec rm -rf []
	xfind . -name "*.go" -exec print {} -debug
*/
import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var EMPTY_PARAMS = make([]string, 0)

var WILL_PROCESS_DIR = false
var WILL_PROCESS_FILE = false

var PATH_MAP = make(map[string]bool)
var DIR_LIST = make([]string, 0)
var FILE_LIST = make([]string, 0)

func WildcardToRegexp(pattern string) (*regexp.Regexp, error) {
	pattern = filepath.ToSlash(pattern) // 转换路径分隔符为/
	pattern = regexp.QuoteMeta(pattern) // 转义特殊字符
	pattern = strings.ReplaceAll(pattern, `\.`, "\\.")
	pattern = strings.ReplaceAll(pattern, `\*`, ".*")
	pattern = strings.ReplaceAll(pattern, `\?`, ".")
	return regexp.Compile("^" + pattern + "$")
}

func MatchWildcard(filename, pattern string) (bool, error) {
	re, err := WildcardToRegexp(pattern)
	if err != nil {
		println("Regexp error.")
		return false, err
	}
	match := re.MatchString(filename)
	return match, nil
}

func fnExclude(path string, info fs.FileInfo) bool {
	for _, exclude := range ARG_EXCLUDES {
		if match, _ := MatchWildcard(info.Name(), exclude); match {
			return true
		}
	}
	return false
}

func matchContent(path string, info fs.FileInfo) bool {

	if len(ARG_MATCHES) > 0 || len(ARG_REGEX) > 0 {

		if info.IsDir() {
			return false
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("printmatch: ERROR opening file:", path)
			return false
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()

			var matched = text
			if matchText(text) || matchRegex(text, &matched) {
				return true
			}
		}
		return false
	}

	return true
}

func fnMatch(path string, info fs.FileInfo) bool {
	if (info.IsDir() && !ARG_TYPE_DIR) || (!info.IsDir() && !ARG_TYPE_FILE) {
		return false
	}

	name := info.Name()
	if len(ARG_NAMES) > 0 {
		match := false
		for _, nameToMatch := range ARG_NAMES {
			if match, _ = MatchWildcard(name, nameToMatch); match {
				break
			}
		}
		if !match {
			return false
		}
	}

	return matchContent(path, info)
}

// return dir, file
func getExecParamType(params []string) (bool, bool) {
	dir := execParamContains(params, "[]")
	file := execParamContains(params, "{}")
	return dir, file
}

// return: willProcessDir, willProcessFile
func checkWillProcess() (bool, bool) {
	willProcessDir := false
	willProcessFile := false

	for _, exec := range ARG_EXEC {
		dir, file := getExecParamType(exec[1:])
		willProcessDir = willProcessDir || dir
		willProcessFile = willProcessFile || file
	}
	return willProcessDir, willProcessFile
}

func fnCallback(path string, info fs.FileInfo) {
	execArgs := preReplaceExecArgs(ARG_EXEC, path, info)

	for _, args := range execArgs {
		cmd := args[0]
		params := args[1:]
		executeArgs(cmd, params, path, info)
	}
	defaultInnerCommand(path, info)
}

func postArgParse() {
	// default to 'both' for type
	if !ARG_TYPE_DIR && !ARG_TYPE_FILE {
		ARG_TYPE_DIR = true
		ARG_TYPE_FILE = true
	}

	// check if dir('[]') or file('{}') param used
	WILL_PROCESS_DIR, WILL_PROCESS_FILE = checkWillProcess()

	// parse inner commands
	parseInnerCommands()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println(`Usage:   xfind [path] [params...]

Parameters:
	-type <type> : 'file', or 'dir', or default 'both'
	-name <filename> : match file name
	-exclude <name> : exclude directory or files
	-match <string> : match file content
	-regex <string> : regex match file content
	-exec <command> : execute command
		internal commands:
		'print' : console output
			e.g. -exec print {} "*" 
			( '{}' for file or directory, '[]' for directory )
		'printmatch' : output matched text line for '-match'
		'count' : count files
		'countlines' : count total file lines
		'countmatch' : count matched file lines
	-debug : display exec command besides execute them
	-delete : delete file
		`)
		return
	}

	path := os.Args[1]
	parseArgs(os.Args[2:])
	postArgParse()

	if err := ScanDirRecursively(path, fnMatch, fnExclude, fnCallback); err != nil {
		fmt.Println("ERROR:", err)
	}

	postInnerCommandExecution()
}

type FnPathMatch func(string, fs.FileInfo) bool
type FnPathExclude func(string, fs.FileInfo) bool
type FnPathCallback func(string, fs.FileInfo)

func ScanDirRecursively(dirPath string, fnMatch FnPathMatch, fnExclude FnPathExclude, fnCallback FnPathCallback) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var absPath = path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(dirPath, path)
			if !filepath.IsAbs(path) {
				absPath, _ = filepath.Abs(absPath)
			}
		}

		if info.IsDir() {
			if fnExclude(absPath, info) {
				return filepath.SkipDir
			} else if fnMatch(absPath, info) {
				fnCallback(absPath, info)
			}
		} else {
			if !fnExclude(absPath, info) && fnMatch(absPath, info) {
				fnCallback(absPath, info)
			}
		}
		return nil
	})
	return err
}
