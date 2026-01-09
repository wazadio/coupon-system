package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerContext struct{}

const (
	LevelDebug = zapcore.DebugLevel
	LevelInfo  = zapcore.InfoLevel
	LevelWarn  = zapcore.WarnLevel
	LevelError = zapcore.ErrorLevel
	LevelFatal = zapcore.FatalLevel
	LevelPanic = zapcore.PanicLevel
)

var Log *zap.Logger

func Init() error {
	return InitWithFile("logs/app.log")
}

// InitWithFile initializes logger with both console and file output
func InitWithFile(logFile string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return err
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create cores for console and file
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Write to both console and file
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapcore.InfoLevel),
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Print(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	if logger, ok := ctx.Value(LoggerContext{}).(*zap.Logger); ok {
		switch level {
		case zapcore.DebugLevel:
			logger.Debug(msg, fields...)
		case zapcore.InfoLevel:
			logger.Info(msg, fields...)
		case zapcore.WarnLevel:
			logger.Warn(msg, fields...)
		case zapcore.ErrorLevel:
			logger.Error(msg, fields...)
		case zapcore.FatalLevel:
			logger.Fatal(msg, fields...)
		case zapcore.PanicLevel:
			logger.Panic(msg, fields...)
		default:
			logger.Info(msg, fields...)
		}
	} else {
		// Fallback to global logger if no contextual logger is found
		switch level {
		case zapcore.DebugLevel:
			Log.Debug(msg, fields...)
		case zapcore.InfoLevel:
			Log.Info(msg, fields...)
		case zapcore.WarnLevel:
			Log.Warn(msg, fields...)
		case zapcore.ErrorLevel:
			Log.Error(msg, fields...)
		case zapcore.FatalLevel:
			Log.Fatal(msg, fields...)
		case zapcore.PanicLevel:
			Log.Panic(msg, fields...)
		default:
			Log.Info(msg, fields...)
		}
	}
}
