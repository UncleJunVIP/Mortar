package common

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var logFile *os.File
var atomicLevel = zap.NewAtomicLevel()
var logger atomic.Pointer[zap.Logger]

var LoggerInitialized atomic.Bool

var onceLogger sync.Once

func LogStandardFatal(msg string, err error) {
	log.SetOutput(os.Stderr)
	log.Fatalf("%s: %v", msg, err)
}

func GetLoggerInstance() *zap.Logger {
	onceLogger.Do(func() {
		logger.Store(createLogger())
	})
	return logger.Load()
}

func CloseLogger() {
	GetLoggerInstance().Sync()
	logFile.Close()
}

func createLogger() *zap.Logger {
	LoggerInitialized.Store(false)

	cwd, err := os.Getwd()
	if err != nil {
		LogStandardFatal("Failed to get current working directory", err)
	}

	logFile, err = os.OpenFile(filepath.Join(cwd, "mortar.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		LogStandardFatal("Unable to open log file!", err)
	}

	writeSyncer := zapcore.AddSync(logFile)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		atomicLevel,
	)

	LoggerInitialized.Store(true)

	return zap.New(core)
}

func SetLogLevel(rawLevel string) {
	var loggerLevel zapcore.Level

	switch strings.ToLower(rawLevel) {
	case "debug":
		loggerLevel = zap.DebugLevel
	case "info":
		loggerLevel = zap.InfoLevel
	case "warn", "warning":
		loggerLevel = zap.WarnLevel
	case "error":
		loggerLevel = zap.ErrorLevel
	case "dpanic":
		loggerLevel = zap.DPanicLevel
	case "panic":
		loggerLevel = zap.PanicLevel
	case "fatal":
		loggerLevel = zap.FatalLevel
	default:
		loggerLevel = zap.InfoLevel
	}

	atomicLevel.SetLevel(loggerLevel)
}
