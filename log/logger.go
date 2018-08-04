package log

import (
	"fmt"
	"os"

	"io"

	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

func init() {
	fmt.Println("init...")
	InitLogToStdoutDebug()
}

func InitLogToStdoutDebug() {
	logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)
}

func InitLogToStdout() {
	logger.Formatter = &logrus.TextFormatter{}
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.WarnLevel)
}

func InitLogToFile(file io.Writer) {
	logger.Formatter = &logrus.JSONFormatter{}
	logger.SetOutput(file)
	logger.SetLevel(logrus.WarnLevel)
}

func Debug(msg string) {
	logger.Debug(msg)
}

// Debugf logs a formatted message with debug log level.
func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

// Info logs a message with info log level.
func Info(msg string) {
	logger.Info(msg)
}

// Infof logs a formatted message with info log level.
func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

// Warn logs a message with warn log level.
func Warn(msg string) {
	logger.Warn(msg)
}

// Warnf logs a formatted message with warn log level.
func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

// Error logs a message with error log level.
func Error(msg string) {
	logger.Error(msg)
}

// Errorf logs a formatted message with error log level.
func Errorf(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

// Fatal logs a message with fatal log level.
func Fatal(msg string) {
	logger.Fatal(msg)
}

// Fatalf logs a formatted message with fatal log level.
func Fatalf(msg string, args ...interface{}) {
	logrus.Fatalf(msg, args...)
}

// Panic logs a message with panic log level.
func Panic(msg string) {
	logger.Panic(msg)
}

// Panicf logs a formatted message with panic log level.
func Panicf(msg string, args ...interface{}) {
	logger.Panicf(msg, args...)
}
