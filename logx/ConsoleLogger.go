package logx

import (
	"fmt"
	"github.com/meiguonet/mgboot-go-common/util/errorx"
)

type consoleLogger struct {
	ch chan []string
}

func NewConsoleLogger(ch chan []string) *consoleLogger {
	return &consoleLogger{ch: ch}
}

func (l *consoleLogger) Log(level interface{}, args ...interface{}) {
	if len(args) < 1 {
		return
	}

	var lv string

	if s1, ok := level.(string); ok && s1 != "" {
		lv = s1
	}

	if lv == "" {
		return
	}

	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			if t != "" {
				l.ch <- []string{lv, t}
			}
		case error:
			if t != nil {
				l.ch <- []string{lv, errorx.Stacktrace(t)}
			}
		}
	}
}

func (l *consoleLogger) Logf(level interface{}, format string, args ...interface{}) {
	var msg string

	if len(args) < 1 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	l.Log(level, msg)
}

func (l *consoleLogger) Trace(args ...interface{}) {
	l.Log("trace", args...)
}

func (l *consoleLogger) Tracef(format string, args ...interface{}) {
	l.Logf("trace", format, args...)
}

func (l *consoleLogger) Debug(args ...interface{}) {
	l.Log("debug", args...)
}

func (l *consoleLogger) Debugf(format string, args ...interface{}) {
	l.Logf("debug", format, args...)
}

func (l *consoleLogger) Info(args ...interface{}) {
	l.Log("info", args...)
}

func (l *consoleLogger) Infof(format string, args ...interface{}) {
	l.Logf("info", format, args...)
}

func (l *consoleLogger) Warn(args ...interface{}) {
	l.Log("warn", args...)
}

func (l *consoleLogger) Warnf(format string, args ...interface{}) {
	l.Logf("warn", format, args...)
}

func (l *consoleLogger) Error(args ...interface{}) {
	l.Log("error", args...)
}

func (l *consoleLogger) Errorf(format string, args ...interface{}) {
	l.Logf("error", format, args...)
}

func (l *consoleLogger) Panic(args ...interface{}) {
	l.Log("panic", args...)
}

func (l *consoleLogger) Panicf(format string, args ...interface{}) {
	l.Logf("panic", format, args...)
}

func (l *consoleLogger) Fatal(args ...interface{}) {
	l.Log("fatal", args...)
}

func (l *consoleLogger) Fatalf(format string, args ...interface{}) {
	l.Logf("fatal", format, args...)
}
