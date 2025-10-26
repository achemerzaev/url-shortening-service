package logger

import "go.uber.org/zap"

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Sync() error
}

type ZapLogger struct {
	*zap.SugaredLogger
}

func New(level string) (Logger, error) {
	var cfg zap.Config
	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	_ = cfg.Level.UnmarshalText([]byte(level))
	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{l.Sugar()}, nil
}
