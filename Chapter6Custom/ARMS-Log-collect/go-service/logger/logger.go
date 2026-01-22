package logger

import (
	"os"
	"time"

	"github.com/go-kit/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	base log.Logger
}

type Config struct {
	Path       string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Env        string
	Version    string
}

func New(cfg Config) *Logger {
	_ = os.MkdirAll("logs", 0755)

	w := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:   cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:    cfg.MaxAgeDays,
		Compress:  false,
	}

	l := log.NewJSONLogger(log.NewSyncWriter(w))
	l = log.With(l,
		"timestamp", log.TimestampFormat(func() time.Time { return time.Now().UTC() }, time.RFC3339Nano),
		"service", "go-service",
		"env", cfg.Env,
		"version", cfg.Version,
	)

	return &Logger{base: l}
}

func (l *Logger) Info(kv ...any)  { _ = l.base.Log(append([]any{"level", "INFO"}, kv...)...) }
func (l *Logger) Warn(kv ...any)  { _ = l.base.Log(append([]any{"level", "WARN"}, kv...)...) }
func (l *Logger) Error(kv ...any) { _ = l.base.Log(append([]any{"level", "ERROR"}, kv...)...) }
func (l *Logger) With(kv ...any) *Logger {
	return &Logger{base: log.With(l.base, kv...)}
}
