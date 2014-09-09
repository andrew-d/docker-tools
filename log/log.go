package log

// Note: much of this is taken from Docker:
//    https://github.com/docker/docker/tree/master/pkg/log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/aybabtme/rgbterm"
)

type priority int

const (
	prefixFormat = "[%s]"
	errorFormat  = "%s %s:%d %s\n"
	logFormat    = "%s %s\n"

	fatalPriority priority = iota
	errorPriority
	warnPriority
	infoPriority
	debugPriority
)

var (
	UseColor bool = true
)

// A common interface to access the Fatal method of
// both testing.B and testing.T.
type Fataler interface {
	Fatal(args ...interface{})
}

func (p priority) String() string {
	switch p {
	case fatalPriority:
		return "fatal"
	case errorPriority:
		return "error"
	case warnPriority:
		return "warn"
	case infoPriority:
		return "info"
	case debugPriority:
		return "debug"
	}

	return ""
}

func (p priority) Colorize(s string) string {
	if !UseColor {
		return s
	}

	switch p {
	case fatalPriority:
		return rgbterm.String(s, 255, 0, 0)
	case errorPriority:
		return rgbterm.String(s, 255, 0, 0)
	case warnPriority:
		return rgbterm.String(s, 255, 255, 0)
	case infoPriority:
		return rgbterm.String(s, 0, 255, 0)
	case debugPriority:
		return rgbterm.String(s, 0, 0, 255)
	}

	return s
}

func Debugf(format string, a ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		logf(os.Stderr, debugPriority, format, a...)
	}
}

func Infof(format string, a ...interface{}) {
	logf(os.Stdout, infoPriority, format, a...)
}

func Warnf(format string, a ...interface{}) {
	logf(os.Stdout, warnPriority, format, a...)
}

func Errorf(format string, a ...interface{}) {
	logf(os.Stderr, errorPriority, format, a...)
}

func Fatalf(format string, a ...interface{}) {
	logf(os.Stderr, fatalPriority, format, a...)
	os.Exit(1)
}

func logf(stream io.Writer, level priority, format string, a ...interface{}) {
	var prefix string

	prefix = fmt.Sprintf(prefixFormat, level.String())

	if level <= errorPriority || level == debugPriority {
		// Retrieve the stack infos
		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "<unknown>"
			line = -1
		} else {
			file = file[strings.LastIndex(file, "/")+1:]
		}

		prefix = fmt.Sprintf(errorFormat, level.Colorize(prefix), file, line, format)
	} else {
		prefix = fmt.Sprintf(logFormat, level.Colorize(prefix), format)
	}

	fmt.Fprintf(stream, prefix, a...)
}
