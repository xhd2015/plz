package logging

import (
	"math"
)

var FallbackLogWriter LogWriter = &stderrLogWriter{}
var LogWriterProviders = []func(loggerKV []interface{}) LogWriter{}
var DefaultLevel = DebugLevel
var CreateLogger = func(loggerKV []interface{}, logWriter LogWriter) Logger {
	return &defaultLogger{loggerKV:loggerKV, logWriter:logWriter, minLevel:DefaultLevel}
}

func LoggerOf(loggerKV ...interface{}) Logger {
	return &placeholder{loggerKV, nil}
}

func realLoggerOf(loggerKV []interface{}) Logger {
	logWriter := getLogWriter(loggerKV)
	return CreateLogger(loggerKV, logWriter)
}

func getLogWriter(loggerKV []interface{}) LogWriter {
	logWriters := []LogWriter{}
	for _, provider := range LogWriterProviders {
		logWriter := provider(loggerKV)
		if logWriter != nil {
			logWriters = append(logWriters, logWriter)
		}
	}
	switch len(logWriters) {
	case 0:
		return FallbackLogWriter
	case 1:
		return logWriters[0]
	default:
		return &combinedLogWriter{logWriters: logWriters}
	}
}

type Level struct {
	Severity  int32
	LevelName string
}

var UndefLevel = Level{math.MaxInt32, ""}
var FatalLevel = Level{60, "FATAL"}
var ErrorLevel = Level{50, "ERROR"}
var WarningLevel = Level{40, "WARNING"}
var InfoLevel = Level{30, "INFO"}
var DebugLevel = Level{20, "DEBUG"}
var TraceLevel = Level{10, "TRACE"}

type Logger interface {
	Log(level Level, msg string, kv ...interface{})
	Error(err error, msg string, kv ...interface{}) error
	Warning(msg string, kv ...interface{})
	Info(msg string, kv ...interface{})
	Debug(msg string, kv ...interface{})
	ShouldLog(level Level) bool
	SetLevel(level Level) Logger
}

type LogWriter interface {
	Log(level Level, msg string, kv []interface{})
}