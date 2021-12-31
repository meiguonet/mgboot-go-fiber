package mgboot

import "github.com/meiguonet/mgboot-go-common/logx"

var runtimeLogger logx.Logger
var requestLogLogger logx.Logger
var logRequestBody bool
var executeTimeLogLogger logx.Logger

func RuntimeLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		runtimeLogger = logger[0]
	}

	l := runtimeLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func RequestLogLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		requestLogLogger = logger[0]
	}

	l := requestLogLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func RequestLogEnabled() bool {
	return requestLogLogger != nil
}

func LogRequestBody(flag ...bool) bool {
	if len(flag) > 0 {
		logRequestBody = flag[0]
	}

	return logRequestBody
}

func ExecuteTimeLogLogger(logger ...logx.Logger) logx.Logger {
	if len(logger) > 0 {
		executeTimeLogLogger = logger[0]
	}

	l := executeTimeLogLogger

	if l == nil {
		l = NewNoopLogger()
	}

	return l
}

func ExecuteTimeLogEnabled() bool {
	return executeTimeLogLogger != nil
}
