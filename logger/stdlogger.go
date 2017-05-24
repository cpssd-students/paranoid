package logger

import (
	"io"
	"os"
)

var stdlog = New("", "", os.DevNull)

func init() {
	stdlog.depth = 4
}

// SetOutput of the default logger
func SetOutput(l LogOutput) {
	stdlog.SetOutput(l)
}

// SetLogDirectory for the standard logger
func SetLogDirectory(dir string) {
	stdlog.logDir = dir
}

// AddAdditionalWriter to the standard log
func AddAdditionalWriter(w io.Writer) {
	stdlog.AddAdditionalWriter(w)
}

// Info logs
func Info(v ...interface{}) {
	stdlog.Info(v...)
}

// Infof logs
func Infof(format string, v ...interface{}) {
	stdlog.Infof(format, v...)
}

// Warn logs
func Warn(v ...interface{}) {
	stdlog.Warn(v...)
}

// Warnf logs
func Warnf(format string, v ...interface{}) {
	stdlog.Warnf(format, v...)
}

// Error logs
func Error(v ...interface{}) {
	stdlog.Error(v...)
}

// Errorf logs
func Errorf(format string, v ...interface{}) {
	stdlog.Errorf(format, v...)
}

// Fatal logs
func Fatal(v ...interface{}) {
	stdlog.Fatal(v...)
}

// Fatalf logs
func Fatalf(format string, v ...interface{}) {
	stdlog.Fatalf(format, v...)
}
