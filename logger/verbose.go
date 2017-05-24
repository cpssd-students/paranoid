package logger

import "flag"

var (
	vFlag = flag.Int("v", 0, "logging verbosity")
)

// V logging verbosity
func V(level int) Verbose {
	// return Verbose(level < *vFlag)
	return Verbose(level <= *vFlag)
}

// Verbose logging
type Verbose bool

// Info logs
func (v Verbose) Info(args ...interface{}) {
	if v {
		stdlog.Info(args...)
	}
}

// Infof logs
func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		stdlog.Infof(format, args...)
	}
}

// Warn logs
func (v Verbose) Warn(args ...interface{}) {
	if v {
		stdlog.Warn(args...)
	}
}

// Warnf logs
func (v Verbose) Warnf(format string, args ...interface{}) {
	if v {
		stdlog.Warnf(format, args...)
	}
}

// Warn logs
func (v Verbose) Error(args ...interface{}) {
	if v {
		stdlog.Error(args...)
	}
}

// Errorf logs
func (v Verbose) Errorf(format string, args ...interface{}) {
	if v {
		stdlog.Errorf(format, args...)
	}
}

// Fatal logs
func (v Verbose) Fatal(args ...interface{}) {
	if v {
		stdlog.Fatal(args...)
	}
}

// Fatalf logs
func (v Verbose) Fatalf(format string, args ...interface{}) {
	if v {
		stdlog.Fatalf(format, args...)
	}
}
