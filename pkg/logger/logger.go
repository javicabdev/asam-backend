package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var log *zap.Logger
var initialized bool

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
)

// Config contiene la configuración del logger
type Config struct {
	// Configuración general
	Level      Level  `json:"level"`
	OutputPath string `json:"output_path"`

	// Configuración de desarrollo
	Development bool `json:"development"`

	// Configuración de rotación de archivos
	MaxSize    int  `json:"max_size"`    // megabytes
	MaxAge     int  `json:"max_age"`     // días
	MaxBackups int  `json:"max_backups"` // número de archivos de backup
	Compress   bool `json:"compress"`    // comprimir logs antiguos

	// Configuración de consola
	ConsoleOutput bool  `json:"console_output"`
	ConsoleLevel  Level `json:"console_level"`
}

// DefaultConfig retorna una configuración por defecto
func DefaultConfig() Config {
	return Config{
		Level:         InfoLevel,
		OutputPath:    "logs/asam.log",
		Development:   false,
		MaxSize:       100,  // 100 MB
		MaxAge:        30,   // 30 días
		MaxBackups:    10,   // 10 archivos de backup
		Compress:      true, // comprimir logs antiguos
		ConsoleOutput: true, // mostrar logs en consola también
		ConsoleLevel:  InfoLevel,
	}
}

// InitLogger inicializa el logger con rotación de archivos
func InitLogger(cfg Config) error {
	// Configurar encoder común
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

	// Crear el core para el archivo con rotación
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

	// Slice para almacenar todos los cores
	var cores []zapcore.Core
	cores = append(cores, fileCore)

	// Agregar salida por consola si está habilitada
	if cfg.ConsoleOutput {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			getZapLevel(cfg.ConsoleLevel),
		)
		cores = append(cores, consoleCore)
	}

	// Crear el logger con todos los cores
	core := zapcore.NewTee(cores...)

	// Agregar las opciones según la configuración
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	// Agregar opción de desarrollo si está habilitada
	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	// Crear el logger con las opciones
	logger := zap.New(core, opts...)

	// Reemplazar el logger global de zap
	zap.ReplaceGlobals(logger)
	log = logger
	initialized = true

	return nil
}

// getZapLevel convierte nuestro Level a zapcore.Level
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
	default:
		return zapcore.InfoLevel
	}
}

// Sync fuerza la escritura de cualquier log en buffer
func Sync() error {
	if !initialized {
		return nil
	}
	return log.Sync()
}

// Métodos para logging

func Debug(msg string, fields ...zap.Field) {
	if !initialized {
		return
	}
	log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	if !initialized {
		return
	}
	log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if !initialized {
		return
	}
	log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if !initialized {
		return
	}
	log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if !initialized {
		os.Exit(1)
	}
	log.Fatal(msg, fields...)
}
