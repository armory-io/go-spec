package logging

import "github.com/sirupsen/logrus"

type Config struct {
	Remote RemoteLoggingConfig `json:"remote" yaml:"remote"`
	JSON   FormatJson          `json:"json" yaml:"json"`
}

type RemoteLoggingConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Version    string `json:"version" yaml:"version"`
	CustomerID string `json:"customerId" yaml:"customerId"`
}

type FormatJson struct {
	Enabled bool              `json:"enabled" yaml:"enabled"`
	Level   string            `json:"level" yaml:"level"`
	Fields  map[string]string `json:"fields" yaml:"fields"`
}

func (fj *FormatJson) Configure(l *logrus.Logger) {
	// TODO - support configuration of field names
	fm := logrus.FieldMap{}
	for f, v := range fj.Fields {
		// we have to switch here because setting fm[f] is of type
		// fieldKey, not a string...
		switch f {
		case logrus.FieldKeyTime:
			fm[logrus.FieldKeyTime] = v
		case logrus.FieldKeyMsg:
			fm[logrus.FieldKeyMsg] = v
		case logrus.FieldKeyLevel:
			fm[logrus.FieldKeyLevel] = v
		}
	}
	formatter := logrus.JSONFormatter{FieldMap: fm}
	l.SetFormatter(&formatter)
	levelOrDefault := func(lvl string, def logrus.Level) logrus.Level {
		logrusLevel, err := logrus.ParseLevel(lvl)
		if err != nil {
			return def
		}
		return logrusLevel
	}

	l.SetLevel(levelOrDefault(fj.Level, logrus.InfoLevel))
}

type LeveledLogger interface {
	WithField(key string, value interface{}) LeveledLogger
	WithFields(fields map[string]interface{}) LeveledLogger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}
