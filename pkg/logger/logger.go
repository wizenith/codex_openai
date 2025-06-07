package logger

import (
	"github.com/sirupsen/logrus"
)

// Init sets the global log level.
func Init(level string) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		l = logrus.InfoLevel
	}
	logrus.SetLevel(l)
}

// Info prints an informational message.
func Info(args ...interface{}) {
	logrus.Info(args...)
}

// Error prints an error message.
func Error(args ...interface{}) {
	logrus.Error(args...)
}
