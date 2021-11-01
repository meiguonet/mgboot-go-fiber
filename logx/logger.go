package logx

import (
	"fmt"
	"github.com/meiguonet/mgboot-go-common/util/errorx"
	"github.com/sirupsen/logrus"
	"strings"
)

type logger struct {
	channel string
	logger  *logrus.Logger
}

func (l *logger) Log(level interface{}, args ...interface{}) {
	logLevel := logrus.TraceLevel

	switch t := level.(type) {
	case logrus.Level:
		logLevel = t
	case string:
		lv := strings.ToLower(t)

		switch strings.ToLower(lv) {
		case "debug":
			logLevel = logrus.DebugLevel
		case "info":
			logLevel = logrus.InfoLevel
		case "warn":
			logLevel = logrus.WarnLevel
		case "error":
			logLevel = logrus.ErrorLevel
		case "panic":
			logLevel = logrus.PanicLevel
		case "fatal":
			logLevel = logrus.FatalLevel
		}
	}

	argList := make([]interface{}, 0)

	for _, v := range args {
		if err, ok := v.(error); ok {
			argList = append(argList, errorx.Stacktrace(err))
			continue
		}

		argList = append(argList, v)
	}

	entry := l.logger.WithField("channel", l.channel)
	entry.Writer()

	switch logLevel {
	case logrus.DebugLevel:
		entry.Debug(argList...)
	case logrus.InfoLevel:
		entry.Info(argList...)
	case logrus.WarnLevel:
		entry.Warn(argList...)
	case logrus.ErrorLevel:
		entry.Error(argList...)
	case logrus.PanicLevel:
		entry.Panic(argList...)
	case logrus.FatalLevel:
		entry.Fatal(argList...)
	default:
		entry.Trace(argList...)
	}
}

func (l *logger) Logf(level interface{}, format string, args ...interface{}) {
	var msg string

	if len(args) < 1 {
		msg = format
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	l.Log(level, msg)
}

func (l *logger) Trace(args ...interface{}) {
	l.Log("trace", args...)
}

func (l *logger) Tracef(format string, args ...interface{}) {
	l.Logf("trace", format, args...)
}

func (l *logger) Debug(args ...interface{}) {
	l.Log("debug", args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.Logf("debug", format, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.Log("info", args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.Logf("info", format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.Log("warn", args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.Logf("warn", format, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.Log("error", args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.Logf("error", format, args...)
}

func (l *logger) Panic(args ...interface{}) {
	l.Log("panic", args...)
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.Logf("panic", format, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.Log("fatal", args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.Logf("fatal", format, args...)
}
