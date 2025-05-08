package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Movemos tus tipos y structs aquí (Config, Level, etc.):

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
	PanicLevel Level = "panic"
)

// Config contiene la configuración del logger
type Config struct {
	Level         Level  `json:"level"`
	OutputPath    string `json:"output_path"`
	Development   bool   `json:"development"`
	MaxSize       int    `json:"max_size"`
	MaxAge        int    `json:"max_age"`
	MaxBackups    int    `json:"max_backups"`
	Compress      bool   `json:"compress"`
	ConsoleOutput bool   `json:"console_output"`
	ConsoleLevel  Level  `json:"console_level"`
}

// DefaultConfig retorna una configuración por defecto
func DefaultConfig() Config {
	return Config{
		Level:         InfoLevel,
		OutputPath:    "logs/asam.log",
		Development:   false,
		MaxSize:       100, // 100 MB
		MaxAge:        30,  // 30 días
		MaxBackups:    10,  // 10 backups
		Compress:      true,
		ConsoleOutput: true,
		ConsoleLevel:  InfoLevel,
	}
}

// zapLogger implementa nuestra interfaz Logger
type zapLogger struct {
	logger *zap.Logger
}

func (zl *zapLogger) Debug(msg string, fields ...zap.Field) {
	zl.logger.Debug(msg, fields...)
}

func (zl *zapLogger) Info(msg string, fields ...zap.Field) {
	zl.logger.Info(msg, fields...)
}

func (zl *zapLogger) Warn(msg string, fields ...zap.Field) {
	zl.logger.Warn(msg, fields...)
}

func (zl *zapLogger) Error(msg string, fields ...zap.Field) {
	zl.logger.Error(msg, fields...)
}

func (zl *zapLogger) Fatal(msg string, fields ...zap.Field) {
	zl.logger.Fatal(msg, fields...)
}

func (zl *zapLogger) Panic(msg string, fields ...zap.Field) {
	zl.logger.Panic(msg, fields...)
}

func (zl *zapLogger) Sync() error {
	return zl.logger.Sync()
}

// InitLogger crea un *zap.Logger según la config, y retorna nuestra interfaz Logger
func InitLogger(cfg Config) (Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}),
		getZapLevel(cfg.Level),
	)

	var cores []zapcore.Core
	cores = append(cores, fileCore)

	if cfg.ConsoleOutput {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			getZapLevel(cfg.ConsoleLevel),
		)
		cores = append(cores, consoleCore)
	}

	core := zapcore.NewTee(cores...)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	zapLog := zap.New(core, opts...)

	// ya no hacemos zap.ReplaceGlobals, ni guardamos en variable global
	return &zapLogger{logger: zapLog}, nil
}

func getZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}
