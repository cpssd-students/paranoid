// +build !integration

package logger

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestInstanceOutput(t *testing.T) {
	l := New("", "", os.DevNull)
	l.SetOutput(NONE)

	var b bytes.Buffer
	l.AddAdditionalWriter(&b)
	l.Info("TestInstanceOutput")

	expected := "I(.{19}) logger_test.go: TestInstanceOutput\n"

	if ok, _ := regexp.Match(expected, b.Bytes()); !ok {
		t.Errorf("expected '%s' got '%s'",
			expected[:len(expected)-1],
			b.String()[:b.Len()-1])
	}
}

func TestInstanceLoggingLevels(t *testing.T) {
	l := New("", "", "")
	l.SetLogLevel(DEBUG)

	cases := []struct {
		exp       rune
		formatted bool
		f         func(v ...interface{})
		ff        func(f string, v ...interface{})
	}{
		{'D', false, l.Debug, nil},
		{'D', true, nil, l.Debugf},
		{'I', false, l.Verbose, nil},
		{'I', true, nil, l.Verbosef},
		{'I', false, l.Info, nil},
		{'I', true, nil, l.Infof},
		{'W', false, l.Warn, nil},
		{'W', true, nil, l.Warnf},
		{'E', false, l.Error, nil},
		{'E', true, nil, l.Errorf},
		// Don't test Fatal... pretty bad idea..
	}

	for i, c := range cases {
		l.SetOutput(NONE)
		var b bytes.Buffer
		l.AddAdditionalWriter(&b)

		if c.formatted {
			c.ff("TestInstanceLoggingLevels %d", i)
		} else {
			c.f("TestInstanceLoggingLevels", i)
		}

		if b.Len() == 0 {
			t.Errorf("c%d: result must not be empty", i)
		}

		if b.Bytes()[0] != byte(c.exp) {
			t.Errorf("c%d: expected the first characted to be %v, got %v",
				i, c.exp, b.String()[0])
		}
	}
}

func TestLoggingLevels(t *testing.T) {
	cases := []struct {
		exp       rune
		formatted bool
		f         func(v ...interface{})
		ff        func(f string, v ...interface{})
	}{
		{'I', false, Info, nil},
		{'I', true, nil, Infof},
		{'W', false, Warn, nil},
		{'W', true, nil, Warnf},
		{'E', false, Error, nil},
		{'E', true, nil, Errorf},
		// Don't test Fatal... pretty bad idea..
	}

	for i, c := range cases {
		SetOutput(NONE)

		var b bytes.Buffer
		AddAdditionalWriter(&b)

		if c.formatted {
			c.ff("TestLoggingLevels %d", i)
		} else {
			c.f("TestLoggingLevels", i)
		}

		if b.Len() == 0 {
			t.Errorf("c%d: result must not be empty", i)
			return
		}

		if b.Bytes()[0] != byte(c.exp) {
			t.Errorf("c%d: expected the first characted to be %v, got %v",
				i, c.exp, b.String()[0])
		}
	}
}

func TestVerboseLoggingLevels(t *testing.T) {
	l := V(0)

	cases := []struct {
		exp       rune
		formatted bool
		f         func(v ...interface{})
		ff        func(f string, v ...interface{})
	}{
		{'I', false, l.Info, nil},
		{'I', true, nil, l.Infof},
		{'W', false, l.Warn, nil},
		{'W', true, nil, l.Warnf},
		{'E', false, l.Error, nil},
		{'E', true, nil, l.Errorf},
		// Don't test Fatal... pretty bad idea..
	}

	for i, c := range cases {
		SetOutput(NONE)
		var b bytes.Buffer
		AddAdditionalWriter(&b)

		if c.formatted {
			c.ff("TestVerboseLoggingLevels %d", i)
		} else {
			c.f("TestVerboseLoggingLevels", i)
		}

		if b.Len() == 0 {
			t.Errorf("c%d: result must not be empty", i)
		}

		if b.Bytes()[0] != byte(c.exp) {
			t.Errorf("c%d: expected the first characted to be %v, got %v",
				i, c.exp, b.String()[0])
		}
	}
}

func TestVerbosity(t *testing.T) {
	cases := []struct {
		logflag  int
		loglevel int
		expected bool
	}{
		{0, 0, true},
		{0, 1, false},
		{0, 2, false},
		{1, 0, true},
		{1, 1, true},
		{1, 2, false},
		{2, 0, true},
		{2, 1, true},
		{2, 2, true},
	}

	for i, c := range cases {
		SetOutput(NONE)

		var b bytes.Buffer
		AddAdditionalWriter(&b)

		flag.Set("-v", fmt.Sprintf("%d", c.logflag))
		V(c.loglevel).Infof("c%d", i)

		yes := "c%d: expected results, have none"
		no := "c%d: expected no results, have %s"

		l := b.Len() > 0
		switch {
		case l && !c.expected:
			t.Errorf(no, i, b.String())
		case !l && c.expected:
			t.Errorf(yes, i)
		}
	}
}

func TestLogFile(t *testing.T) {
	testPath := "/tmp/paranoid"
	defer os.RemoveAll(testPath)
	SetLogDirectory(testPath)
	// defer os.RemoveAll(testPath)
	SetOutput(LOGFILE)

	Info("TestLogFile")

	data, err := ioutil.ReadFile(filepath.Join(testPath, "logger.log"))
	if err != nil {
		t.Errorf("error reading file: %v", err)
	}

	expected := "I(.{19}) logger_test.go: TestLogFile\n"

	if ok, _ := regexp.Match(expected, data); !ok {
		t.Errorf("expected '%s' got '%s'",
			expected[:len(expected)-1],
			string(data)[:len(data)-1])
	}
}
