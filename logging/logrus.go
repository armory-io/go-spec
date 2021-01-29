package logging

import "github.com/sirupsen/logrus"

type LogrusAdapter struct {
	*logrus.Entry
}

func (l *LogrusAdapter) WithField(key string, value interface{}) LeveledLogger {
	withField := l.Logger.WithField(key, value)
	return &LogrusAdapter{withField}
}

func (l *LogrusAdapter) WithFields(fields map[string]interface{}) LeveledLogger {
	panic("implement me")
}
