package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	// Stdout is exported for tests.
	Stdout io.Writer = os.Stdout
	// Stderr is exported for tests.
	Stderr io.Writer = os.Stderr
	// LogWriter is exposed for consuming extra command output, if needed.
	LogWriter io.Writer = io.Discard
)

var opts struct {
	logger     *log.Logger
	funcs      map[string]bool
	files      map[string]bool
	logFile    *os.File
	logSecrets bool
	verbosity  int
}

// v is a verbosity level.
type v int

// V returns a logger at the given verbosity level.
// The higher the number, the more verbose the logging.
func V(n int) v {
	return v(n)
}

var logFn = doNotLog

// make sure all initializations happens before the init func.
var enabled = initDebug()

func initDebug() bool {
	if opts.logFile != nil {
		_ = opts.logFile.Close()
	}

	if l := os.Getenv("GOPASS_DEBUG_VERBOSE"); l != "" {
		if iv, err := strconv.Atoi(l); err == nil {
			opts.verbosity = iv
		}
	}

	if os.Getenv("GOPASS_DEBUG") == "" && os.Getenv("GOPASS_DEBUG_LOG") == "" {
		logFn = doNotLog

		return false
	}

	// we need to explicitly set logSecrets to false in case tests run under an environment
	// where GOPASS_DEBUG_LOG_SECRETS is true. Otherwise setting it to false in the test
	// context won't have any effect.
	opts.logSecrets = false
	if sv := os.Getenv("GOPASS_DEBUG_LOG_SECRETS"); sv != "" && sv != "false" {
		opts.logSecrets = true
	}

	initDebugLogger()
	initDebugTags()

	logFn = doLog

	return true
}

func initDebugLogger() {
	debugfile := os.Getenv("GOPASS_DEBUG_LOG")
	if debugfile == "" {
		opts.logger = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
		LogWriter = os.Stderr

		return
	}

	f, err := os.OpenFile(debugfile, os.O_WRONLY|os.O_APPEND, 0o600)
	if err == nil {
		// seek to the end of the file (offset, whence [2 = end])
		_, err := f.Seek(0, 2)
		if err != nil {
			_ = f.Close()
			fmt.Fprintf(Stderr, "unable to seek to end of %v: %v\n", debugfile, err)
			os.Exit(3)
		}
	}

	if err != nil && os.IsNotExist(err) {
		f, err = os.OpenFile(debugfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	}

	if err != nil {
		fmt.Fprintf(Stderr, "unable to open debug log file %v: %v\n", debugfile, err)
		os.Exit(2)
	}

	opts.logFile = f
	opts.logger = log.New(f, "", log.Ldate|log.Lmicroseconds)
	LogWriter = f
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

		switch t[0] {
		case '-':
			val = false
			t = t[1:]
		case '+':
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
		s += ":*"
	}

	return s
}

func initDebugTags() {
	opts.funcs = parseFilter("GOPASS_DEBUG_FUNCS", padFunc)
	opts.files = parseFilter("GOPASS_DEBUG_FILES", padFile)
}

func getPosition(offset int) (fn, dir, file string, line int) { //nolint:nonamedreturns
	pc, file, line, ok := runtime.Caller(3 + offset)
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
// debug log file (if enabled), but only if the verbosity
// level is greater or equal to the given level.
//
// This is a no-op if the verbosity level is not high enough.
func (n v) Log(f string, args ...any) {
	logFn(int(n), 0, f, args...)
}

// Log logs a statement to Stderr (unless filtered) and the
// debug log file (if enabled).
//
// This is a no-op if debugging is not enabled.
func Log(f string, args ...any) {
	logFn(0, 0, f, args...)
}

// LogN logs a statement to Stderr (unless filtered) and the
// debug log file (if enabled). The offset will be applied to
// the runtime position.
//
// This is a no-op if debugging is not enabled.
func LogN(offset int, f string, args ...any) {
	logFn(0, offset, f, args...)
}

func doNotLog(verbosity, offset int, f string, args ...any) {}

func doLog(verbosity, offset int, f string, args ...any) {
	// if the log message is too verbose for the requested verbosity level, skip it
	if verbosity > opts.verbosity {
		return
	}

	fn, dir, file, line := getPosition(offset)

	if len(f) == 0 || f[len(f)-1] != '\n' {
		f += "\n"
	}

	type Shortener interface {
		Str() string
	}

	type Safer interface {
		SafeStr() string
	}

	argsi := make([]any, len(args))
	for i, item := range args {
		argsi[i] = item
		if secreter, ok := item.(Safer); ok && !opts.logSecrets {
			argsi[i] = secreter.SafeStr()

			continue
		}

		if shortener, ok := item.(Shortener); ok {
			argsi[i] = shortener.Str()
		}
	}

	pos := fmt.Sprintf("%s/%s:%d", dir, file, line)

	formatString := fmt.Sprintf("%s\t%s\t%s", pos, fn, f)

	dbgprint := func() {
		fmt.Fprintf(Stderr, formatString, argsi...)
	}

	if opts.logger != nil {
		opts.logger.Printf(formatString, argsi...)
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

// IsEnabled returns true if debug logging was enabled.
// This is useful to avoid expensive computations if debugging is not enabled.
func IsEnabled() bool {
	return enabled
}
