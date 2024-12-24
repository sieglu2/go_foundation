package foundation

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// globalLogger is global zap logger.
	globalLogger Logging = NewSugarLogger(getLogLevel())
)

func getLogLevel() string {
	levelFromEnv := os.Getenv("LOG_LEVEL")
	if len(levelFromEnv) == 0 {
		levelFromEnv = "info"
	}
	return levelFromEnv
}

type Logging interface {
	With(args ...interface{}) Logging
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

func Logger() Logging {
	return globalLogger
}

func OverrideGlobalLogger(logging Logging) {
	globalLogger = logging
}

func NewSugarLogger(lvlStr string) *SugarLogger {
	// First, define our level-handling logic.
	globalLevel, err := zapcore.ParseLevel(lvlStr)
	if err != nil {
		log.Fatalf("failed to initialize global logger: %v", err)
	}

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	// It is useful for Kubernetes deployment.
	// Kubernetes interprets os.Stdout log items as INFO and os.Stderr log items
	// as ERROR by default.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= globalLevel && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Configure console output.
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(cfg)

	// Join the outputs, encoders, and level-handling functions into zapcore
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	zapLogger := zap.New(core)
	zap.RedirectStdLog(zapLogger)

	return &SugarLogger{
		internal: zapLogger.Sugar(),
	}
}

func NewEmptyLogger() *EmptyLogger {
	return &EmptyLogger{}
}

type SugarLogger struct {
	internal *zap.SugaredLogger
}

func (s *SugarLogger) With(args ...interface{}) Logging {
	return &SugarLogger{
		internal: s.internal.With(args...),
	}
}

func (s *SugarLogger) Debug(args ...interface{}) {
	s.internal.Debug(args...)
}

func (s *SugarLogger) Info(args ...interface{}) {
	s.internal.Info(args...)
}

func (s *SugarLogger) Warn(args ...interface{}) {
	s.internal.Warn(append(args, "\nCallstack:\n", GetCallStack()))
}

func (s *SugarLogger) Error(args ...interface{}) {
	s.internal.Error(append(args, "\nCallstack:\n", GetCallStack()))
}

func (s *SugarLogger) Fatal(args ...interface{}) {
	s.internal.Fatal(append(args, "\nCallstack:\n", GetCallStack()))
}

func (s *SugarLogger) Debugf(template string, args ...interface{}) {
	s.internal.Debugf(template, args...)
}

func (s *SugarLogger) Infof(template string, args ...interface{}) {
	s.internal.Infof(template, args...)
}

func (s *SugarLogger) Warnf(template string, args ...interface{}) {
	s.internal.Warn(appendCallstack(template, args...))
}

func (s *SugarLogger) Errorf(template string, args ...interface{}) {
	s.internal.Error(appendCallstack(template, args...))
}

func (s *SugarLogger) Fatalf(template string, args ...interface{}) {
	s.internal.Fatal(appendCallstack(template, args...))
}

type EmptyLogger struct {
}

func (s *EmptyLogger) With(args ...interface{}) Logging {
	return &EmptyLogger{}
}

func (s *EmptyLogger) Debug(args ...interface{}) {
}

func (s *EmptyLogger) Info(args ...interface{}) {
}

func (s *EmptyLogger) Warn(args ...interface{}) {
}

func (s *EmptyLogger) Error(args ...interface{}) {
}

func (s *EmptyLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func (s *EmptyLogger) Debugf(template string, args ...interface{}) {
}

func (s *EmptyLogger) Infof(template string, args ...interface{}) {
}

func (s *EmptyLogger) Warnf(template string, args ...interface{}) {
}

func (s *EmptyLogger) Errorf(template string, args ...interface{}) {
}

func (s *EmptyLogger) Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}

func GetCallStack() string {
	return string(debug.Stack())
}

func appendCallstack(template string, args ...interface{}) string {
	liquidated := fmt.Sprintf(template, args...)
	stack := strings.ReplaceAll(GetCallStack(), `\n`, "\n")
	return fmt.Sprintf("%s\nCallstack:\n%s", liquidated, stack)
}
