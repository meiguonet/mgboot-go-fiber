package logx

import (
	clogx "github.com/meiguonet/mgboot-go-common/logx"
	"github.com/meiguonet/mgboot-go-common/util/fsx"
	"github.com/meiguonet/mgboot-go-fiber"
	"os"
)

var logDir string
var loggers = map[string]clogx.Logger{}

func WithLogDir(dir string) {
	dir = fsx.GetRealpath(dir)
	
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		logDir = dir
	}
}

func GetLogDir() string {
	return logDir
}

func WithLogger(name string, logger clogx.Logger) {
	loggers[name] = logger
}

func Channel(name string) clogx.Logger {
	logger := loggers[name]
	
	if logger == nil {
		logger = mgboot.NewNoopLogger()
	}
	
	return logger
}

func Log(level interface{}, args ...interface{}) {
	mgboot.RuntimeLogger().Log(level, args...)
}

func Logf(level interface{}, format string, args ...interface{}) {
	mgboot.RuntimeLogger().Logf(level, format, args...)
}

func Trace(args ...interface{}) {
	mgboot.RuntimeLogger().Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Tracef(format, args...)
}

func Debug(args ...interface{}) {
	mgboot.RuntimeLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	mgboot.RuntimeLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	mgboot.RuntimeLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	mgboot.RuntimeLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Errorf(format, args...)
}

func Panic(args ...interface{}) {
	mgboot.RuntimeLogger().Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Panicf(format, args...)
}

func Fatal(args ...interface{}) {
	mgboot.RuntimeLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	mgboot.RuntimeLogger().Infof(format, args...)
}
