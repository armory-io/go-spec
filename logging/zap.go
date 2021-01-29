package logging

import "go.uber.org/zap"

type ZapAdapter struct {
	*zap.SugaredLogger
}

func (z *ZapAdapter) WithField(key string, value interface{}) LeveledLogger {
	fz := z.SugaredLogger.With(key, value)
	return &ZapAdapter{fz}
}

func (z *ZapAdapter) WithFields(fields map[string]interface{}) LeveledLogger {
	fz := z.SugaredLogger
	for k, v := range fields {
		fz = fz.With(k, v)
	}
	return &ZapAdapter{fz}
}

func NewZapLeveledLogger() (LeveledLogger, error) {
	l, _ := zap.NewProduction()
	sl := l.Sugar()
	return &ZapAdapter{sl}, nil
}
