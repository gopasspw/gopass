package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	// Stdout is exported for tests
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests
	Stderr io.Writer = os.Stderr
)

var opts struct {
	logger *log.Logger
	funcs  map[string]bool
	files  map[string]bool
}

var logFn = doNotLog

// make sure all initializations happens before the init func
var enabled = initDebug()

func initDebug() bool {
	if os.Getenv("GOPASS_DEBUG") == "" && os.Getenv("GOPASS_DEBUG_LOG") == "" {
		logFn = doNotLog
		return false
	}

	initDebugLogger()
	initDebugTags()
	logFn = doLog

	return true
}

func initDebugLogger() {
	debugfile := os.Getenv("GOPASS_DEBUG_LOG")
	if debugfile == "" {
		return
	}

	f, err := os.OpenFile(debugfile, os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		_, err := f.Seek(2, 0)
		if err != nil {
			fmt.Fprintf(Stderr, "unable to seek to end of %v: %v\n", debugfile, err)
			os.Exit(3)
		}
	}
	if err != nil && os.IsNotExist(err) {
		f, err = os.OpenFile(debugfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	}

	if err != nil {
		fmt.Fprintf(Stderr, "unable to open debug log file %v: %v\n", debugfile, err)
		os.Exit(2)
	}

	opts.logger = log.New(f, "", log.LstdFlags)
}

func parseFilter(envname string, pad func(string) string) map[string]bool {
	filter := make(map[string]bool)

	env := os.Getenv(envname)
	if env == "" {
		return filter
	}

	for _, fn := range strings.Split(env, ",") {
		t := pad(strings.TrimSpace(fn))
		val := true
		if t[0] == '-' {
			val = false
			t = t[1:]
		} else if t[0] == '+' {
			val = true
			t = t[1:]
		}

		// test pattern
		_, err := path.Match(t, "")
		if err != nil {
			fmt.Fprintf(Stderr, "error: invalid pattern %q: %v\n", t, err)
			os.Exit(5)
		}

		filter[t] = val
	}

	return filter
}

func padFunc(s string) string {
	if s == "all" {
		return s
	}

	return s
}

func padFile(s string) string {
	if s == "all" {
		return s
	}

	if !strings.Contains(s, "/") {
		s = "*/" + s
	}

	if !strings.Contains(s, ":") {
		s = s + ":*"
	}

	return s
}

func initDebugTags() {
	opts.funcs = parseFilter("GOPASS_DEBUG_FUNCS", padFunc)
	opts.files = parseFilter("GOPASS_DEBUG_FILES", padFile)
}

func getPosition() (fn, dir, file string, line int) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "", "", "", 0
	}

	dirname := filepath.Base(filepath.Dir(file))
	filename := filepath.Base(file)

	f := runtime.FuncForPC(pc)
	return path.Base(f.Name()), dirname, filename, line
}

func checkFilter(filter map[string]bool, key string) bool {
	// check if exact match
	if v, ok := filter[key]; ok {
		return v
	}

	// check globbing
	for k, v := range filter {
		if m, _ := path.Match(k, key); m {
			return v
		}
	}

	// check if tag "all" is enabled
	if v, ok := filter["all"]; ok && v {
		return true
	}

	return false
}

// Log logs a statement to Stderr (unless filtered) and the
// debug log file (if enabled).
func Log(f string, args ...interface{}) {
	logFn(f, args...)
}

func doNotLog(f string, args ...interface{}) {}

func doLog(f string, args ...interface{}) {
	fn, dir, file, line := getPosition()
	if len(f) == 0 || f[len(f)-1] != '\n' {
		f += "\n"
	}

	type Shortener interface {
		Str() string
	}

	for i, item := range args {
		if shortener, ok := item.(Shortener); ok {
			args[i] = shortener.Str()
		}
	}

	pos := fmt.Sprintf("%s/%s:%d", dir, file, line)

	formatString := fmt.Sprintf("%s\t%s\t%s", pos, fn, f)

	dbgprint := func() {
		fmt.Fprintf(Stderr, formatString, args...)
	}

	if opts.logger != nil {
		opts.logger.Printf(formatString, args...)
	}

	filename := fmt.Sprintf("%s/%s:%d", dir, file, line)
	if checkFilter(opts.files, filename) {
		dbgprint()
		return
	}

	if checkFilter(opts.funcs, fn) {
		dbgprint()
	}
}

// IsEnabled returns true if debug logging was enabled
func IsEnabled() bool {
	return enabled
}
