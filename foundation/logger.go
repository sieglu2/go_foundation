package foundation

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogColorRed    = "\033[31m"
	LogColorYellow = "\033[33m"
	LogColorReset  = "\033[0m"
)

var (
	globalLogger Logging = NewSugarLogger(getLogLevel(), getLogFormat())
)

func getLogLevel() string {
	levelFromEnv := os.Getenv("ZAP_LOG_LEVEL")
	if len(levelFromEnv) == 0 {
		levelFromEnv = "info"
	}
	return strings.ToLower(levelFromEnv)
}

func getLogFormat() string {
	logFormatFromEnv := os.Getenv("ZAP_LOG_FORMAT")
	if len(logFormatFromEnv) == 0 {
		logFormatFromEnv = "console"
	}
	return strings.ToLower(logFormatFromEnv)
}

type Logging interface {
	With(args ...any) Logging

	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)

	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
}

func Logger() Logging {
	return globalLogger
}

func LoadGlobalLogger(logging Logging) {
	globalLogger = logging
}

func NewSimpleLogger(lvlStr string) *SimpleLogger {
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			// Don't encode level
		},
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			// Don't encode timestamp
		},
		EncodeDuration: zapcore.StringDurationEncoder,
	})

	return &SimpleLogger{
		internal: newZapLogger(lvlStr, consoleEncoder),
	}
}

func NewSugarLogger(lvlStr string, logFormat string) *SugarLogger {
	var consoleEncoder zapcore.Encoder

	switch logFormat {
	case "json":
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = zapcore.RFC3339TimeEncoder
		consoleEncoder = zapcore.NewJSONEncoder(cfg)

	case "console":
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeLevel = zapcore.CapitalLevelEncoder
		cfg.EncodeTime = zapcore.RFC3339TimeEncoder
		consoleEncoder = zapcore.NewConsoleEncoder(cfg)

	default:
		log.Fatalf("Unsupported log format: %s", logFormat)
	}

	return &SugarLogger{
		internal: newZapLogger(lvlStr, consoleEncoder),
	}
}

func newZapLogger(lvlStr string, encoder zapcore.Encoder) *zap.SugaredLogger {
	// First, define our level-handling logic.
	targetLevel, err := zapcore.ParseLevel(lvlStr)
	if err != nil {
		log.Fatalf("failed to initialize global logger: %v", err)
	}

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= targetLevel && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Join the outputs, encoders, and level-handling functions into zapcore
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, consoleErrors, highPriority),
		zapcore.NewCore(encoder, consoleInfos, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	zapLogger := zap.New(core)
	zap.RedirectStdLog(zapLogger)

	return zapLogger.Sugar()
}

// SimpleLogger
type SimpleLogger struct {
	internal *zap.SugaredLogger
}

func (s *SimpleLogger) With(args ...any) Logging {
	return &SimpleLogger{
		internal: s.internal.With(args...),
	}
}

func (s *SimpleLogger) Debug(args ...any) {
	s.internal.Debug(args...)
}

func (s *SimpleLogger) Info(args ...any) {
	s.internal.Info(args...)
}

func (s *SimpleLogger) Warn(args ...any) {
	s.internal.Warn(args...)
}

func (s *SimpleLogger) Error(args ...any) {
	s.internal.Error(args...)
}

func (s *SimpleLogger) Fatal(args ...any) {
	s.internal.Fatal(append(args, "\nCallstack:\n", GetCallStack()))
}

func (s *SimpleLogger) Debugf(template string, args ...any) {
	s.internal.Debugf(template, args...)
}

func (s *SimpleLogger) Infof(template string, args ...any) {
	s.internal.Infof(template, args...)
}

func (s *SimpleLogger) Warnf(template string, args ...any) {
	s.internal.Warnf(template, args...)
}

func (s *SimpleLogger) Errorf(template string, args ...any) {
	s.internal.Errorf(template, args...)
}

func (s *SimpleLogger) Fatalf(template string, args ...any) {
	s.internal.Fatalf(template, args...)
}

// SugarLogger
type SugarLogger struct {
	internal *zap.SugaredLogger
}

func (s *SugarLogger) With(args ...any) Logging {
	return &SugarLogger{
		internal: s.internal.With(args...),
	}
}

func (s *SugarLogger) Debug(args ...any) {
	s.internal.Debug(args...)
}

func (s *SugarLogger) Info(args ...any) {
	s.internal.Info(args...)
}

func (s *SugarLogger) Warn(args ...any) {
	finalMsg := fmt.Sprintf("%v\nCallstack:\n%s", fmt.Sprint(args...), GetCallStack())
	s.internal.Warn(ColorizeMessage(LogColorYellow, finalMsg))
}

func (s *SugarLogger) Error(args ...any) {
	finalMsg := fmt.Sprintf("%v\nCallstack:\n%s", fmt.Sprint(args...), GetCallStack())
	s.internal.Error(ColorizeMessage(LogColorRed, finalMsg))
}

func (s *SugarLogger) Fatal(args ...any) {
	finalMsg := fmt.Sprintf("%v\nCallstack:\n%s", fmt.Sprint(args...), GetCallStack())
	s.internal.Fatal(ColorizeMessage(LogColorRed, finalMsg))
}

func (s *SugarLogger) Debugf(template string, args ...any) {
	s.internal.Debugf(template, args...)
}

func (s *SugarLogger) Infof(template string, args ...any) {
	s.internal.Infof(template, args...)
}

func (s *SugarLogger) Warnf(template string, args ...any) {
	finalMsg := appendCallstack(template, args...)
	s.internal.Warn(ColorizeMessage(LogColorYellow, finalMsg))
}

func (s *SugarLogger) Errorf(template string, args ...any) {
	finalMsg := appendCallstack(template, args...)
	s.internal.Error(ColorizeMessage(LogColorRed, finalMsg))
}

func (s *SugarLogger) Fatalf(template string, args ...any) {
	finalMsg := appendCallstack(template, args...)
	s.internal.Fatal(ColorizeMessage(LogColorRed, finalMsg))
}

func GetCallStack() string {
	return string(debug.Stack())
}

func appendCallstack(template string, args ...any) string {
	liquidated := fmt.Sprintf(template, args...)
	stack := strings.ReplaceAll(GetCallStack(), `\n`, "\n")
	return fmt.Sprintf("%s\nCallstack:\n%s", liquidated, stack)
}

func ColorizeMessage(color, message string) string {
	return fmt.Sprintf("%s%s%s", color, message, LogColorReset)
}
