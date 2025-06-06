package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	dryRun bool
}

func NewLogger(level zap.AtomicLevel) *Logger {
	cfg := zap.Config{
		Encoding:         "console",
		Level:            level,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.RFC3339TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logger, _ := cfg.Build()
	return &Logger{Logger: logger}
}

func (l *Logger) EnableDryRunLogger() {
	l.dryRun = true
}

func (l *Logger) DryRunLogger() *zap.Logger {
	if !l.dryRun {
		return zap.NewNop() // suppress output when not dry-run
	}
	return l.With(zap.String("loggingMode", "dry-run"))
}

func (l *Logger) CloneWith(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.With(fields...),
		dryRun: l.dryRun,
	}
}
