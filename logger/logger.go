package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger            *Logger
	loggerDebugWriter *LoggerWriter
	loggerInfoWriter  *LoggerWriter
	loggerWarnWriter  *LoggerWriter
	loggerErrorWriter *LoggerWriter
)

func RegisterLoggerWriter(name, level string, writer io.Writer) {
	switch {
	case level == "debug" && loggerDebugWriter != nil:
		loggerDebugWriter.registeryWriter[name] = writer
	case level == "info" && loggerInfoWriter != nil:
		loggerInfoWriter.registeryWriter[name] = writer
	case level == "warn" && loggerWarnWriter != nil:
		loggerWarnWriter.registeryWriter[name] = writer
	case level == "error" && loggerErrorWriter != nil:
		loggerErrorWriter.registeryWriter[name] = writer
	default:
		if loggerDebugWriter != nil {
			loggerDebugWriter.registeryWriter[name] = writer
		}
		if loggerInfoWriter != nil {
			loggerInfoWriter.registeryWriter[name] = writer
		}
		if loggerWarnWriter != nil {
			loggerWarnWriter.registeryWriter[name] = writer
		}
		if loggerErrorWriter != nil {
			loggerErrorWriter.registeryWriter[name] = writer
		}
	}
}

func UnregisterLoggerWriter(name string) {
	if loggerDebugWriter != nil {
		delete(loggerDebugWriter.registeryWriter, name)
	}
	if loggerInfoWriter != nil {
		delete(loggerInfoWriter.registeryWriter, name)
	}
	if loggerWarnWriter != nil {
		delete(loggerWarnWriter.registeryWriter, name)
	}
	if loggerErrorWriter != nil {
		delete(loggerErrorWriter.registeryWriter, name)
	}
}

type LoggerWriter struct {
	zapcore.WriteSyncer
	registeryWriter map[string]io.Writer
}

func NewLoggerWriter(writer io.Writer) *LoggerWriter {
	return &LoggerWriter{WriteSyncer: zapcore.AddSync(writer), registeryWriter: make(map[string]io.Writer)}
}

func (w *LoggerWriter) Write(p []byte) (n int, err error) {
	for _, writer := range w.registeryWriter {
		_, _ = writer.Write(p)
	}
	return w.WriteSyncer.Write(p)
}

func init() {
	checkLogger()
}

func checkLogger() {
	if logger == nil {
		InitTestLogger()
	}
}

type FormaterLogger interface {
	Infof(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
}

type StdLogger interface {
	Info(msg string, args ...zap.Field)
	Debug(msg string, args ...zap.Field)
	Warn(msg string, args ...zap.Field)
	Error(msg string, args ...zap.Field)
	Panic(msg string, args ...zap.Field)
	Fatal(msg string, args ...zap.Field)
}

type ILogger interface {
	FormaterLogger
	StdLogger
	Wrap(name string) any
}

func GetLogger() *Logger {
	return logger
}

func Infof(template string, args ...interface{}) {
	logger.l.Info(fmt.Sprintf(template, args...))
}
func Debugf(template string, args ...interface{}) {
	logger.l.Debug(fmt.Sprintf(template, args...))
}
func Warnf(template string, args ...interface{}) {
	logger.l.Warn(fmt.Sprintf(template, args...))
}
func Errorf(template string, args ...interface{}) {
	logger.l.Error(fmt.Sprintf(template, args...))
}
func Panicf(template string, args ...interface{}) {
	logger.l.Panic(fmt.Sprintf(template, args...))
}
func Info(msg string, args ...zap.Field) {
	logger.l.Info(msg, args...)
}
func Debug(msg string, args ...zap.Field) {
	logger.l.Debug(msg, args...)
}
func Warn(msg string, args ...zap.Field) {
	logger.l.Warn(msg, args...)
}
func Error(msg string, args ...zap.Field) {
	logger.l.Error(msg, args...)
}
func Panic(msg string, args ...zap.Field) {
	logger.l.Panic(msg, args...)
}
func Fatal(msg string, args ...zap.Field) {
	logger.l.Fatal(msg, args...)
}

// Sync 同步流数据
//
//	写入文件时不是立刻就写入的，若要停止程序，需要在停止程序前调用，以保证日志可以完全写入到文件中
func Sync() {
	_ = logger.l.Sync()
}

type LogConfig struct {
	Name        string
	LogPath     string
	Level       zapcore.Level
	MaxSize     int
	MaxBackup   int
	MaxAge      int
	HideConsole bool
}

func InitLogger(logConfig LogConfig) {
	logger = newLogger(logConfig)
}

func InitConsoleLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02-15:04:05.000")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeCaller = nil
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(config), os.Stdout, zap.DebugLevel)
	logger = &Logger{
		l: zap.New(consoleCore, zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}
func InitTestLogger() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02-15:04:05.000")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCore := zapcore.NewCore(zapcore.NewConsoleEncoder(config), os.Stdout, zap.DebugLevel)
	logger = &Logger{
		l: zap.New(consoleCore, zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

type Logger struct {
	l *zap.Logger
}

func newLogger(logConfig LogConfig) *Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02-15:04:05.000")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(config)
	var cores []zapcore.Core
	if !logConfig.HideConsole {
		cores = append(cores, zapcore.NewCore(encoder, os.Stdout, logConfig.Level))
	} else {
		errFile, err := os.OpenFile(filepath.Join(logConfig.LogPath, logConfig.Name+"_stderr.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		_, _ = errFile.WriteString(fmt.Sprintf("==========================> pid: %d <========================== %s\n", os.Getpid(), time.Now().Format(time.DateTime)))
		outFile, err := os.OpenFile(filepath.Join(logConfig.LogPath, logConfig.Name+"_stdout.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		_, _ = outFile.WriteString(fmt.Sprintf("==========================> pid: %d <========================== %s\n", os.Getpid(), time.Now().Format(time.DateTime)))
		redirectStderr(outFile, errFile)
	}
	if logConfig.LogPath != "" {
		_ = os.Mkdir(logConfig.LogPath, 0644)
		cores = append(cores,
			debugFileCore(encoder, logConfig.LogPath, logConfig.Name),
			infoFileCore(encoder, logConfig.LogPath, logConfig.Name),
			errorFileCore(encoder, logConfig.LogPath, logConfig.Name),
			warnFileCore(encoder, logConfig.LogPath, logConfig.Name),
		)
	}
	core := zapcore.NewTee(cores...)
	return &Logger{
		l: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.IncreaseLevel(logConfig.Level)),
	}
}
func infoFileCore(consoleEncoder zapcore.Encoder, logPath, filename string) zapcore.Core {
	loggerInfoWriter = NewLoggerWriter(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, filename+"_info.log"),
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	})
	return zapcore.NewCore(consoleEncoder,
		loggerInfoWriter,
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == zapcore.InfoLevel
		}))
}
func debugFileCore(consoleEncoder zapcore.Encoder, logPath, filename string) zapcore.Core {
	loggerDebugWriter = NewLoggerWriter(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, filename+"_debug.log"),
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	})
	return zapcore.NewCore(consoleEncoder,
		loggerDebugWriter,
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == zapcore.DebugLevel
		}))
}
func errorFileCore(consoleEncoder zapcore.Encoder, logPath, filename string) zapcore.Core {
	loggerErrorWriter = NewLoggerWriter(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, filename+"_error.log"),
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	})
	return zapcore.NewCore(consoleEncoder,
		loggerErrorWriter,
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level >= zapcore.ErrorLevel
		}))
}

func warnFileCore(consoleEncoder zapcore.Encoder, logPath, filename string) zapcore.Core {
	loggerWarnWriter = NewLoggerWriter(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, filename+"_warn.log"),
		MaxSize:    10,
		MaxBackups: 7,
		MaxAge:     7,
		Compress:   true,
	})
	return zapcore.NewCore(consoleEncoder,
		loggerWarnWriter,
		zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == zapcore.WarnLevel
		}))
}
func (l *Logger) Wrap(name string) any {
	n := *l
	n.l = l.l.With(zap.Namespace(name))
	return &n
}
func (l *Logger) Info(msg string, args ...zap.Field) {
	l.l.Info(msg, args...)
}
func (l *Logger) Debug(msg string, args ...zap.Field) {
	l.l.Debug(msg, args...)
}
func (l *Logger) Warn(msg string, args ...zap.Field) {
	l.l.Warn(msg, args...)
}
func (l *Logger) Error(msg string, args ...zap.Field) {
	l.l.Error(msg, args...)
}
func (l *Logger) Panic(msg string, args ...zap.Field) {
	l.l.Panic(msg, args...)
}
func (l *Logger) Fatal(msg string, args ...zap.Field) {
	l.l.Fatal(msg, args...)
}
func (l *Logger) Infof(template string, args ...interface{}) {
	l.l.Info(fmt.Sprintf(template, args...))
}
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.l.Debug(fmt.Sprintf(template, args...))
}
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.l.Warn(fmt.Sprintf(template, args...))
}
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.l.Error(fmt.Sprintf(template, args...))
}
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.l.Panic(fmt.Sprintf(template, args...))
}
