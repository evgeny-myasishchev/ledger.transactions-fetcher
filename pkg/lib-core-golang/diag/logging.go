package diag

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

type contextKey string

const loggerKey contextKey = "logger"

// MsgData - represents msgData structure
type MsgData map[string]interface{}

// Logger - logger interface
type Logger interface {
	// TODO: Support methods below should be actually Somethingf
	// so should be renamed

	Error(ctx context.Context, msg string, args ...interface{})
	Warn(ctx context.Context, msg string, args ...interface{})
	Info(ctx context.Context, msg string, args ...interface{})
	Debug(ctx context.Context, msg string, args ...interface{})

	WithError(err error) Logger
	WithData(data MsgData) Logger
}

type logrusLogger struct {
	target *logrus.Logger
	entry  *logrus.Entry
}

func newLogrusLogger(out io.Writer) logrusLogger {
	logger := logrusLogger{
		target: &logrus.Logger{
			Out:       out,
			Formatter: new(logrus.JSONFormatter),
			Level:     logrus.DebugLevel,
		},
	}
	return *logger.withField("v", 1)
}

type logrusTarget interface {
	Log(level logrus.Level, args ...interface{})
	WithError(err error) *logrus.Entry
	WithTime(t time.Time) *logrus.Entry
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
}

func (logger *logrusLogger) getTarget() logrusTarget {
	if logger.entry != nil {
		return logger.entry
	}
	return logger.target
}

func (logger *logrusLogger) log(ctx context.Context, level logrus.Level, msg string, args ...interface{}) {
	target := logger.getTarget()
	if ctx != nil {
		requestID := RequestIDValue(ctx)
		if requestID != "" {
			target = target.WithField("context", map[string]string{"requestID": requestID})
		}
	}

	if len(args) > 0 {
		target.Log(level, fmt.Sprintf(msg, args...))
	} else {
		target.Log(level, msg)
	}
}

func (logger *logrusLogger) WithError(err error) Logger {
	entry := logger.getTarget().WithError(err)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithTime(t time.Time) Logger {
	entry := logger.getTarget().WithTime(t)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) withField(key string, value interface{}) *logrusLogger {
	entry := logger.getTarget().WithField(key, value)
	childLogger := &logrusLogger{
		target: logger.target,
		entry:  entry,
	}
	return childLogger
}

func (logger *logrusLogger) WithData(data MsgData) Logger {
	return logger.withField("msgData", data)
}

func (logger *logrusLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	logger.log(ctx, logrus.ErrorLevel, msg, args...)
}

func (logger *logrusLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	logger.log(ctx, logrus.InfoLevel, msg, args...)
}

func (logger *logrusLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	logger.log(ctx, logrus.WarnLevel, msg, args...)
}

func (logger *logrusLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	logger.log(ctx, logrus.DebugLevel, msg, args...)
}

// LoggingSystemSetup - logging system setup interface
type LoggingSystemSetup interface {
	SetLogMode(string)
	SetLogLevel(string)
}

type loggingSystem struct {
	logger      logrusLogger
	projectRoot string
}

func (s *loggingSystem) SetLogMode(mode string) {
	switch mode {
	case "json":
		s.logger.target.Formatter = new(logrus.JSONFormatter)
	case "test":
		path := filepath.Join(s.projectRoot, "test.log")
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		s.logger.target.Out = file
	}
}

/* SetLogLevel sets min level to output. Possible values:
- error
- warn
- info
- debug
*/
func (s *loggingSystem) SetLogLevel(level string) {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	s.logger.target.Level = logrusLevel
}

var defaultLoggingSystem loggingSystem

func init() {
	if _, file, _, ok := runtime.Caller(0); ok == true {
		defaultLoggingSystem.projectRoot = filepath.Join(file, "..", "..", "..", "..")
	} else {
		panic("Can not get project root")
	}

	defaultLoggingSystem.logger = newLogrusLogger(os.Stdout)

	if v := flag.Lookup("test.v"); v == nil {
		defaultLoggingSystem.SetLogMode("json")
	} else {
		defaultLoggingSystem.SetLogMode("test")
	}
}

// SetupLoggingSystem initializes a root logger that is a base for all other loggers
// This method should be called just once during APP bootstrap
func SetupLoggingSystem(setup ...func(LoggingSystemSetup)) {
	for _, setupFn := range setup {
		setupFn(&defaultLoggingSystem)
	}
}

// CreateLogger will return logger derived from a rootLogger
// This is suitable for module wide logger
func CreateLogger() Logger {
	loggerName := "unknown"
	if _, file, _, ok := runtime.Caller(1); ok == true {
		loggerName = filepath.Dir(file)
	}
	loggerName, err := filepath.Rel(defaultLoggingSystem.projectRoot, loggerName)
	if err != nil {
		fmt.Printf("Failed to resolve relative path. Base: %v, relative: %v\n", defaultLoggingSystem.projectRoot, loggerName)
		fmt.Println(err)
	}
	return defaultLoggingSystem.logger.withField("package", loggerName)
}
