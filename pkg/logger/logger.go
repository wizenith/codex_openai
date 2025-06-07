package logger

import (
	"log"
)

// Info prints an informational message.
func Info(args ...interface{}) {
	log.Println(args...)
}

// Error prints an error message.
func Error(args ...interface{}) {
	log.Println(args...)
}
