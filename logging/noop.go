package logging

// NoopLeveledLogger implements the LeveledLogger interface
// and throws away all output
type NoopLeveledLogger struct{}

func (n *NoopLeveledLogger) WithField(key string, value interface{}) LeveledLogger {
	return n
}

func (n *NoopLeveledLogger) WithFields(fields map[string]interface{}) LeveledLogger {
	return n
}

func (n *NoopLeveledLogger) Debugf(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Infof(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Warnf(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Errorf(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Fatalf(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Panicf(format string, args ...interface{}) {
}

func (n *NoopLeveledLogger) Debug(args ...interface{}) {
}

func (n *NoopLeveledLogger) Info(args ...interface{}) {
}

func (n *NoopLeveledLogger) Warn(args ...interface{}) {
}

func (n *NoopLeveledLogger) Error(args ...interface{}) {
}

func (n *NoopLeveledLogger) Fatal(args ...interface{}) {
}

func (n *NoopLeveledLogger) Panic(args ...interface{}) {
}
