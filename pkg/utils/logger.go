package utils

import (
	"NATS_TIRE_SERVICE/internal/config"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(cfg *config.Config) (*Logger, error) {
	var level zapcore.Level

	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Создаем базовую конфигурацию энкодера
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Если нужно консольное форматирование - переопределяем
	if cfg.LogFormat == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig() // Заменяем, а не создаем новую
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	}

	var cores []zapcore.Core

	stdoutSyncer := zapcore.Lock(os.Stdout)
	var consoleEncoder zapcore.Encoder

	// Создаем энкодер на основе исправленной конфигурации
	if cfg.LogFormat == "json" {
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	cores = append(cores, zapcore.NewCore(consoleEncoder, stdoutSyncer, level))

	if cfg.LogFile != "" {
		fileSyncer, err := getLogFileSyncer(cfg.LogFile)
		if err != nil {
			return nil, err
		}
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, level))
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel))

	logger = logger.With(
		zap.String("service", cfg.AppName),
		zap.String("version", cfg.Version),
		zap.String("env", cfg.Env))

	return &Logger{logger}, nil
}

func getLogFileSyncer(logFile string) (zapcore.WriteSyncer, error) {
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return zapcore.AddSync(file), nil
}

func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) LogDuration(start time.Time, operation string) {
	duration := time.Since(start)
	l.Debug("operation completed",
		zap.String("operation", operation),
		zap.Duration("duration", duration))
}

// Debugf логирует отладочное сообщение с форматированием
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug("", zap.String("message", fmt.Sprintf(format, args...)))
}

// Infof логирует информационное сообщение с форматированием
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info("", zap.String("message", fmt.Sprintf(format, args...)))
}

// Warnf логирует предупреждение с форматированием
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn("", zap.String("message", fmt.Sprintf(format, args...)))
}

// Errorf логирует ошибку с форматированием
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error("", zap.String("message", fmt.Sprintf(format, args...)))
}

// Fatalf логирует фатальную ошибку с форматированием
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Fatal("", zap.String("message", fmt.Sprintf(format, args...)))
}
