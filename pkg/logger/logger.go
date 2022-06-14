package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

// LogLevel is an abstraction over int that allows to better undestand the
// input of SetLogLevel
type LogLevel int

// Different logging levels
const (
	DEBUG LogLevel = iota
	VERBOSE
	INFO
	WARNING
	ERROR
)

// LogLevel aliases
const (
	WARN  = WARNING
	FATAL = ERROR // In terms of logging importance they are equivelent
)

var levels = []string{
	"D",
	"V",
	"I",
	"W",
	"E",
	"F",
}

func (l LogLevel) String() string {
	return levels[l]
}

// LogOutput is an enum to set the outputs
type LogOutput int

// Enums representing different log locations.
const (
	NONE    LogOutput = 1 << (iota + 1)
	STDERR  LogOutput = 1 << (iota + 1)
	LOGFILE LogOutput = 1 << (iota + 1)
)

const (
	// Package of the project
	Package = "paranoid/"
	timeFmt = "2006-01-02T15:04.05"
)

// ParanoidLogger struct containing the variables necessary for the logger
type ParanoidLogger struct {
	depth     int
	logDir    string
	writers   map[string]io.Writer
	writer    io.Writer
	logLevel  LogLevel
	logToFile bool
}

// New creates a new logger and returns a new logger
// WARNING: currentPackage and component variables will not be used
func New(currentPackage string, component string, logDirectory string) *ParanoidLogger {
	return &ParanoidLogger{
		depth:    3,
		logDir:   logDirectory,
		writers:  map[string]io.Writer{"stderr": os.Stderr},
		writer:   ioutil.Discard,
		logLevel: INFO,
	}
}

// SetLogLevel sets the logging level for the logger
func (l *ParanoidLogger) SetLogLevel(level LogLevel) {
	l.logLevel = level
}

// SetOutput sets the default output for the logger
// TODO: remove error
func (l *ParanoidLogger) SetOutput(output LogOutput) error {
	defer l.refreshWriters()
	if NONE&output == NONE {
		l.writers = make(map[string]io.Writer)
		return nil
	}
	if LOGFILE&output == LOGFILE {
		l.logToFile = true
	}
	if STDERR&output == STDERR {
		l.writers["stderr"] = os.Stderr
		return nil
	}
	// If stderr was not in the list remove it.
	delete(l.writers, "stderr")
	return nil
}

// AddAdditionalWriter allows to add a custom writer to the logger.
// This can be cleared by calling logger.SetOutput() again
func (l *ParanoidLogger) AddAdditionalWriter(w io.Writer) {
	defer l.refreshWriters()
	if _, ok := l.writers["custom"]; !ok {
		l.writers["custom"] = w
		return
	}
	l.writers["custom"] = io.MultiWriter(l.writers["custom"], w)
}

func (l *ParanoidLogger) refreshWriters() {
	for _, w := range l.writers {
		l.writer = io.MultiWriter(l.writer, w)
	}
}

///////////////////////////////// DEBUG /////////////////////////////////

// Debug only prints if LogLevel is set to DEBUG
func (l *ParanoidLogger) Debug(v ...interface{}) {
	if l.logLevel <= DEBUG {
		l.output(DEBUG, v...)
	}
}

// Debugf only prints if LogLevel is set to DEBUG
func (l *ParanoidLogger) Debugf(format string, v ...interface{}) {
	if l.logLevel <= DEBUG {
		l.outputf(DEBUG, format, v...)
	}
}

///////////////////////////////// VERBOSE /////////////////////////////////

// Verbose only prints if LogLevel is set to VERBOSE or lower in importance
func (l *ParanoidLogger) Verbose(v ...interface{}) {
	if l.logLevel <= VERBOSE {
		l.output(INFO, v...)
	}
}

// Verbosef only prints if LogLevel is set to VERBOSE or lower in importance
func (l *ParanoidLogger) Verbosef(format string, v ...interface{}) {
	if l.logLevel <= VERBOSE {
		l.outputf(INFO, format, v...)
	}
}

///////////////////////////////// INFO /////////////////////////////////

// Info only prints if LogLevel is set to INFO or lower in importance
func (l *ParanoidLogger) Info(v ...interface{}) {
	if l.logLevel <= INFO {
		l.output(INFO, v...)
	}
}

// Infof only prints if LogLevel is set to INFO or lower in importance
func (l *ParanoidLogger) Infof(format string, v ...interface{}) {
	if l.logLevel <= INFO {
		l.outputf(INFO, format, v...)
	}
}

///////////////////////////////// WARN /////////////////////////////////

// Warn only prints if LogLevel is set to WARNING or lower in importance
func (l *ParanoidLogger) Warn(v ...interface{}) {
	if l.logLevel <= WARNING {
		l.output(WARNING, v...)
	}
}

// Warnf only prints if LogLevel is set to WARNING or lower in importance
func (l *ParanoidLogger) Warnf(format string, v ...interface{}) {
	if l.logLevel <= WARNING {
		l.outputf(WARNING, format, v...)
	}
}

///////////////////////////////// ERROR /////////////////////////////////

// Error only prints if LogLevel is set to ERROR or lower in importance
func (l *ParanoidLogger) Error(v ...interface{}) {
	if l.logLevel <= ERROR {
		l.output(ERROR, v...)
	}
}

// Errorf only prints if LogLevel is set to ERROR or lower in importance
func (l *ParanoidLogger) Errorf(format string, v ...interface{}) {
	if l.logLevel <= ERROR {
		l.outputf(ERROR, format, v...)
	}
}

///////////////////////////////// FATAL /////////////////////////////////

// Fatal always prints and exits the program with exit code 1
func (l *ParanoidLogger) Fatal(v ...interface{}) {
	l.output(FATAL, v...)
	os.Exit(1)
}

// Fatalf always prints and exits the program with exit code 1
func (l *ParanoidLogger) Fatalf(format string, v ...interface{}) {
	l.outputf(FATAL, format, v...)
	os.Exit(1)
}

///////////////////////////////// GENERAL /////////////////////////////////

func (l *ParanoidLogger) output(level LogLevel, v ...interface{}) {
	l.write(level, v...)
}

func (l *ParanoidLogger) outputf(level LogLevel, format string, v ...interface{}) {
	l.write(level, fmt.Sprintf(format, v...))
}

func (l *ParanoidLogger) write(level LogLevel, v ...interface{}) {
	// Get the dir of the package that is executing the logging.
	_, file, _, _ := runtime.Caller(l.depth)
	suffix := strings.Split(file, Package)[1]
	splitPkg := strings.Split(suffix, "/")
	component := splitPkg[0]

	if l.logToFile {
		if _, ok := l.writers["comp_"+component]; !ok {
			// TODO: figure out what to do with the dismissed error
			var err error
			l.writers["comp_"+component], err = createFileWriter(l.logDir, component)
			if err != nil {
				panic(err)
			}
			l.refreshWriters()
		}
	}

	pkg := strings.Join(splitPkg[1:], "/")

	now := time.Now().Format(timeFmt)
	out := fmt.Sprintf("%s%s %s: %s\n", level.String(), now, pkg, fmt.Sprint(v...))
	l.writer.Write([]byte(out))
}

func createFileWriter(logPath string, component string) (io.Writer, error) {
	os.MkdirAll(logPath, 0700)
	return os.OpenFile(
		path.Join(logPath, component+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
}
