package mgboot

type noopLogger struct {
}

func NewNoopLogger() *noopLogger {
	return &noopLogger{}
}

func (l *noopLogger) Log(_ interface{}, _ ...interface{}) {
}

func (l *noopLogger) Logf(_ interface{}, _ string, _ ...interface{}) {
}

func (l *noopLogger) Trace(_ ...interface{}) {
}

func (l *noopLogger) Tracef(_ string, _ ...interface{}) {
}

func (l *noopLogger) Debug(_ ...interface{}) {
}

func (l *noopLogger) Debugf(_ string, _ ...interface{}) {
}

func (l *noopLogger) Info(_ ...interface{}) {
}

func (l *noopLogger) Infof(_ string, _ ...interface{}) {
}

func (l *noopLogger) Warn(_ ...interface{}) {
}

func (l *noopLogger) Warnf(_ string, _ ...interface{}) {
}

func (l *noopLogger) Error(_ ...interface{}) {
}

func (l *noopLogger) Errorf(_ string, _ ...interface{}) {
}

func (l *noopLogger) Panic(_ ...interface{}) {
}

func (l *noopLogger) Panicf(_ string, _ ...interface{}) {
}

func (l *noopLogger) Fatal(_ ...interface{}) {
}

func (l *noopLogger) Fatalf(_ string, _ ...interface{}) {
}
